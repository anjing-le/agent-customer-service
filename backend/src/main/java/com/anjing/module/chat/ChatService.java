package com.anjing.module.chat;

import com.anjing.module.agent.application.AgentRuntime;
import com.anjing.module.agent.domain.AgentReply;
import com.anjing.module.agent.domain.ConversationMessage;
import com.anjing.module.agent.domain.ConversationRole;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.KnowledgeEvidence;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.KnowledgeSource;
import com.anjing.module.agent.domain.PromptRenderResult;
import com.anjing.module.agent.domain.ReasoningStep;
import com.anjing.module.agent.domain.RuleHit;
import com.anjing.module.chat.entity.ChatAgentAudit;
import com.anjing.module.chat.entity.ChatMessage;
import com.anjing.module.chat.entity.ChatRuntimeSnapshot;
import com.anjing.module.chat.entity.ChatSession;
import com.anjing.module.chat.entity.ChatTransferTicket;
import com.anjing.module.chat.repository.ChatAgentAuditRepository;
import com.anjing.module.chat.repository.ChatMessageRepository;
import com.anjing.module.chat.repository.ChatRuntimeSnapshotRepository;
import com.anjing.module.chat.repository.ChatSessionRepository;
import com.anjing.module.chat.repository.ChatTransferTicketRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.PageRequest;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * 对话中心服务。
 *
 * 当前类只保留会话/消息持久化和前端 VO 映射，可靠 Agent 编排由 AgentRuntime 承载。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class ChatService {

    private static final int RUNTIME_HISTORY_LIMIT = 10;

    private final ChatSessionRepository sessionRepository;
    private final ChatMessageRepository messageRepository;
    private final ChatAgentAuditRepository auditRepository;
    private final ChatRuntimeSnapshotRepository snapshotRepository;
    private final ChatTransferTicketRepository transferTicketRepository;
    private final AgentRuntime agentRuntime;

    /**
     * 获取对话运行概览。
     */
    public ChatVO.RuntimeOverviewVO getRuntimeOverview() {
        LocalDateTime startOfDay = LocalDate.now().atStartOfDay();
        LocalDateTime now = LocalDateTime.now();

        long totalSessions = sessionRepository.count();
        long totalMessages = messageRepository.count();

        ChatVO.RuntimeOverviewVO vo = new ChatVO.RuntimeOverviewVO();
        vo.setTotalSessions(totalSessions);
        vo.setActiveSessions(sessionRepository.countByStatus("active"));
        vo.setTodaySessions(sessionRepository.countByCreatedAtBetween(startOfDay, now));
        vo.setTotalMessages(totalMessages);
        vo.setTodayMessages(messageRepository.countByCreatedAtBetween(startOfDay, now));
        vo.setTodayUserMessages(messageRepository.countByRoleAndCreatedAtBetween("user", startOfDay, now));
        vo.setTodayAssistantMessages(messageRepository.countByRoleAndCreatedAtBetween("assistant", startOfDay, now));
        vo.setTodayAgentReplies(auditRepository.countByCreatedAtBetween(startOfDay, now));
        vo.setTodaySafeReplies(auditRepository.countBySafeAndCreatedAtBetween(true, startOfDay, now));
        vo.setTodayFallbackReplies(auditRepository.countByFallbackRequiredAndCreatedAtBetween(true, startOfDay, now));
        vo.setAverageMessagesPerSession(totalSessions == 0 ? 0D : Math.round((double) totalMessages * 100 / totalSessions) / 100D);
        vo.setRecentSessions(loadRecentSessions());
        vo.setRecentAudits(loadRecentAudits());
        vo.setQualitySummary(buildQualitySummary());
        vo.setDailyTrends(buildDailyTrends());
        vo.setLatestSnapshot(loadLatestSnapshot());
        vo.setTransferSummary(buildTransferSummary(startOfDay, now));
        vo.setRecentTransferTickets(loadRecentTransferTickets());
        return vo;
    }

    /**
     * 采样当前运行态，沉淀为历史快照。
     */
    @Transactional
    public ChatVO.RuntimeSnapshotVO captureRuntimeSnapshot(String snapshotType) {
        ChatVO.RuntimeOverviewVO overview = getRuntimeOverview();
        ChatVO.QualitySummaryVO quality = overview.getQualitySummary();

        ChatRuntimeSnapshot snapshot = new ChatRuntimeSnapshot();
        snapshot.setSnapshotId(UUID.randomUUID().toString());
        snapshot.setSnapshotDate(LocalDate.now());
        snapshot.setSnapshotType(snapshotType);
        snapshot.setTotalSessions(overview.getTotalSessions());
        snapshot.setActiveSessions(overview.getActiveSessions());
        snapshot.setTotalMessages(overview.getTotalMessages());
        snapshot.setTotalAuditedReplies(quality.getTotalAuditedReplies());
        snapshot.setAverageConfidence(quality.getAverageConfidence());
        snapshot.setFallbackRate(quality.getFallbackRate());
        snapshot.setUnsafeRate(quality.getUnsafeRate());
        snapshot.setAverageKnowledgeEvidenceCount(quality.getAverageKnowledgeEvidenceCount());
        snapshot.setAverageRuleHitCount(quality.getAverageRuleHitCount());
        snapshot.setAveragePromptRenderCount(quality.getAveragePromptRenderCount());
        return toRuntimeSnapshotVO(snapshotRepository.save(snapshot));
    }

    /**
     * 创建会话
     */
    @Transactional
    public ChatVO.SessionVO createSession(ChatDTO.CreateSessionDTO dto) {
        String sessionId = UUID.randomUUID().toString().replace("-", "");

        ChatSession session = new ChatSession();
        session.setSessionId(sessionId);
        session.setUserId(dto.getUserId());
        session.setUserName(dto.getUserName());
        session.setChannel(dto.getChannel() != null ? dto.getChannel() : "web");
        session.setStatus("active");
        sessionRepository.save(session);

        ChatMessage welcomeMsg = new ChatMessage();
        welcomeMsg.setMessageId(UUID.randomUUID().toString());
        welcomeMsg.setSessionId(sessionId);
        welcomeMsg.setRole("assistant");
        welcomeMsg.setContent("您好！我是智能客服小助手，请问有什么可以帮助您的吗？");
        messageRepository.save(welcomeMsg);

        ChatVO.SessionVO vo = new ChatVO.SessionVO();
        vo.setSessionId(sessionId);
        vo.setUserId(dto.getUserId());
        vo.setUserName(dto.getUserName());
        vo.setStatus("active");
        vo.setCreatedAt(session.getCreatedAt());
        vo.setMessageCount(1);

        log.info("创建会话成功: sessionId={}", sessionId);
        return vo;
    }

    /**
     * 获取会话列表
     */
    public List<ChatVO.SessionVO> listSessions(ChatDTO.QuerySessionDTO dto) {
        List<ChatSession> sessions;

        if (dto.getUserId() != null && dto.getStatus() != null) {
            sessions = sessionRepository.findByUserIdAndStatusOrderByCreatedAtDesc(dto.getUserId(), dto.getStatus());
        } else if (dto.getUserId() != null) {
            sessions = sessionRepository.findByUserIdOrderByCreatedAtDesc(dto.getUserId());
        } else if (dto.getStatus() != null) {
            sessions = sessionRepository.findByStatusOrderByCreatedAtDesc(dto.getStatus());
        } else {
            sessions = sessionRepository.findAllByOrderByCreatedAtDesc();
        }

        List<ChatVO.SessionVO> result = new ArrayList<>();
        for (ChatSession session : sessions) {
            ChatVO.SessionVO vo = new ChatVO.SessionVO();
            vo.setSessionId(session.getSessionId());
            vo.setUserId(session.getUserId());
            vo.setUserName(session.getUserName());
            vo.setStatus(session.getStatus());
            vo.setCreatedAt(session.getCreatedAt());
            vo.setMessageCount((int) messageRepository.countBySessionId(session.getSessionId()));

            List<ChatMessage> messages = messageRepository.findBySessionIdOrderByCreatedAtAsc(session.getSessionId());
            if (!messages.isEmpty()) {
                vo.setLastMessage(messages.get(messages.size() - 1).getContent());
            }
            result.add(vo);
        }
        return result;
    }

    /**
     * 获取会话详情
     */
    public ChatVO.SessionDetailVO getSession(String sessionId) {
        return sessionRepository.findById(sessionId).map(session -> {
            ChatVO.SessionDetailVO detail = new ChatVO.SessionDetailVO();
            detail.setSessionId(session.getSessionId());
            detail.setUserId(session.getUserId());
            detail.setUserName(session.getUserName());
            detail.setStatus(session.getStatus());
            detail.setCreatedAt(session.getCreatedAt());

            List<ChatMessage> messages = messageRepository.findBySessionIdOrderByCreatedAtAsc(sessionId);
            detail.setMessages(messages.stream().map(this::toMessageVO).toList());
            detail.setContext(new ChatVO.ContextVO());
            detail.setSessionQuality(buildSessionQuality(sessionId));
            detail.setSessionAudits(loadSessionAudits(sessionId));
            detail.setLatestTransferTicket(loadLatestTransferTicket(sessionId));
            detail.setTransferTickets(listTransferTicketsBySession(sessionId));
            return detail;
        }).orElse(null);
    }

    /**
     * 删除会话
     */
    @Transactional
    public void deleteSession(String sessionId) {
        messageRepository.deleteBySessionId(sessionId);
        sessionRepository.deleteById(sessionId);
        log.info("删除会话成功: sessionId={}", sessionId);
    }

    /**
     * 发送消息。
     */
    @Transactional
    public ChatVO.SendMessageVO sendMessage(ChatDTO.SendMessageDTO dto) {
        String sessionId = dto.getSessionId();
        log.info("对话消息进入 AgentRuntime: sessionId={}, content={}", sessionId, dto.getContent());

        ChatMessage userMessage = new ChatMessage();
        userMessage.setMessageId(UUID.randomUUID().toString());
        userMessage.setSessionId(sessionId);
        userMessage.setRole("user");
        userMessage.setContent(dto.getContent());
        messageRepository.save(userMessage);

        ConversationTurn turn = buildConversationTurn(dto);
        AgentReply agentReply = agentRuntime.handle(turn);

        ChatMessage aiMessage = new ChatMessage();
        aiMessage.setMessageId(UUID.randomUUID().toString());
        aiMessage.setSessionId(sessionId);
        aiMessage.setRole("assistant");
        aiMessage.setContent(agentReply.getContent());
        aiMessage.setCardType(agentReply.getCardType());
        messageRepository.save(aiMessage);

        agentReply.setMessageId(aiMessage.getMessageId());
        saveAgentAudit(dto.getSessionId(), aiMessage, agentReply);
        ChatVO.TransferTicketVO transferTicket = ensureTransferTicket(sessionId, aiMessage, agentReply);
        return toSendMessageVO(agentReply, aiMessage, buildSessionQuality(sessionId), transferTicket);
    }

    /**
     * 获取消息历史
     */
    public List<ChatVO.MessageVO> getMessages(ChatDTO.QueryMessagesDTO dto) {
        int page = dto.getPage() != null ? dto.getPage() - 1 : 0;
        int size = dto.getSize() != null ? dto.getSize() : 20;

        List<ChatMessage> messages = messageRepository.findBySessionIdOrderByCreatedAtAsc(
                dto.getSessionId(), PageRequest.of(page, size));
        return messages.stream().map(this::toMessageVO).toList();
    }

    /**
     * 更新上下文（V2 runtime 先从请求 extra 读取，持久化上下文后续接 Redis/JPA）。
     */
    public void updateContext(ChatDTO.UpdateContextDTO dto) {
        log.info("更新上下文: sessionId={}", dto.getSessionId());
    }

    /**
     * 获取上下文
     */
    public ChatVO.ContextVO getContext(String sessionId) {
        return new ChatVO.ContextVO();
    }

    /**
     * 查询转人工工单。
     */
    public List<ChatVO.TransferTicketVO> listTransferTickets(ChatDTO.QueryTransferTicketDTO dto) {
        List<ChatTransferTicket> tickets;
        if (dto.getSessionId() != null && !dto.getSessionId().isBlank()) {
            tickets = transferTicketRepository.findBySessionIdOrderByCreatedAtDesc(dto.getSessionId());
        } else {
            tickets = transferTicketRepository.findAll();
        }
        return tickets.stream()
                .filter(ticket -> dto.getStatus() == null || dto.getStatus().isBlank() || dto.getStatus().equals(ticket.getStatus()))
                .sorted((left, right) -> right.getCreatedAt().compareTo(left.getCreatedAt()))
                .map(this::toTransferTicketVO)
                .toList();
    }

    /**
     * 模拟人工接管完成并回写结果。
     */
    @Transactional
    public ChatVO.TransferTicketVO resolveTransferTicket(ChatDTO.ResolveTransferTicketDTO dto) {
        ChatTransferTicket ticket = transferTicketRepository.findById(dto.getTicketId()).orElse(null);
        if (ticket == null) {
            return null;
        }
        ticket.setStatus("RESOLVED");
        ticket.setAssignedAgentId(dto.getAgentId() != null ? dto.getAgentId() : "agent_demo");
        ticket.setAssignedAgentName(dto.getAgentName() != null ? dto.getAgentName() : "演示坐席");
        ticket.setResolutionNote(dto.getResolutionNote() != null ? dto.getResolutionNote() : "人工已接管并完成处理");
        ticket.setResolvedAt(LocalDateTime.now());
        return toTransferTicketVO(transferTicketRepository.save(ticket));
    }

    private ConversationTurn buildConversationTurn(ChatDTO.SendMessageDTO dto) {
        ConversationTurn turn = new ConversationTurn();
        turn.setTurnId(UUID.randomUUID().toString());
        turn.setSessionId(dto.getSessionId());
        turn.setUserMessage(dto.getContent());
        turn.setRecentHistory(loadRecentHistory(dto.getSessionId()));
        turn.setContext(buildRuntimeContext(dto));
        return turn;
    }

    private List<ConversationMessage> loadRecentHistory(String sessionId) {
        List<ChatMessage> messages = messageRepository.findBySessionIdOrderByCreatedAtAsc(sessionId);
        int historySize = Math.max(0, messages.size() - 1);
        int start = Math.max(0, historySize - RUNTIME_HISTORY_LIMIT);

        List<ConversationMessage> history = new ArrayList<>();
        for (int i = start; i < historySize; i++) {
            history.add(toConversationMessage(messages.get(i)));
        }
        log.info("AgentRuntime 多轮上下文: 历史消息{}条", history.size());
        return history;
    }

    private Map<String, Object> buildRuntimeContext(ChatDTO.SendMessageDTO dto) {
        Map<String, Object> context = new HashMap<>();
        if (dto.getExtra() != null) {
            context.putAll(dto.getExtra());
        }
        context.put("messageType", dto.getMessageType() != null ? dto.getMessageType() : "text");
        return context;
    }

    private ChatVO.SendMessageVO toSendMessageVO(
            AgentReply agentReply,
            ChatMessage aiMessage,
            ChatVO.SessionQualityVO sessionQuality,
            ChatVO.TransferTicketVO transferTicket
    ) {
        ChatVO.SendMessageVO response = new ChatVO.SendMessageVO();
        response.setMessageId(aiMessage.getMessageId());
        response.setContent(agentReply.getContent());
        response.setCardType(agentReply.getCardType());
        response.setSceneType(agentReply.getIntentAnalysis().getSceneType());
        response.setIntent(toIntentVO(agentReply));
        response.setEmotion(agentReply.getIntentAnalysis().getEmotion());
        response.setKnowledgeRecall(toKnowledgeRecallVO(agentReply.getKnowledgeRecall()));
        response.setReasoningProcess(toReasoningProcess(agentReply.getReasoningSteps()));
        response.setReliability(toReliabilityVO(agentReply));
        response.setSessionQuality(sessionQuality);
        response.setSessionAudits(loadSessionAudits(aiMessage.getSessionId()));
        response.setLatestTransferTicket(transferTicket);
        response.setCreatedAt(aiMessage.getCreatedAt());
        return response;
    }

    private ChatVO.ReliabilityVO toReliabilityVO(AgentReply agentReply) {
        ChatVO.ReliabilityVO vo = new ChatVO.ReliabilityVO();
        vo.setReplyEngine(agentReply.getEngine() != null ? agentReply.getEngine().name() : null);
        if (agentReply.getGuardrailDecision() != null) {
            vo.setSafe(agentReply.getGuardrailDecision().isSafe());
            vo.setFallbackRequired(agentReply.getGuardrailDecision().isFallbackRequired());
            vo.setFallbackReason(agentReply.getGuardrailDecision().getFallbackReason() != null
                    ? agentReply.getGuardrailDecision().getFallbackReason().name()
                    : null);
            vo.setUserVisibleNotice(agentReply.getGuardrailDecision().getUserVisibleNotice());
            vo.setTransferRecommended(agentReply.getGuardrailDecision().isTransferRecommended());
            vo.setTransferReason(agentReply.getGuardrailDecision().getTransferReason());
            vo.setTransferPriority(agentReply.getGuardrailDecision().getTransferPriority());
            vo.setPolicyTags(agentReply.getGuardrailDecision().getPolicyTags());
            vo.setRuleHits(toRuleHitVOs(agentReply.getGuardrailDecision().getRuleHits()));
        }
        vo.setPromptRenders(toPromptRenderVOs(agentReply.getPromptRenderResults()));
        return vo;
    }

    private List<ChatVO.RuleHitVO> toRuleHitVOs(List<RuleHit> ruleHits) {
        if (ruleHits == null) return List.of();
        return ruleHits.stream().map(hit -> {
            ChatVO.RuleHitVO vo = new ChatVO.RuleHitVO();
            vo.setRuleCode(hit.getRuleCode());
            vo.setRuleName(hit.getRuleName());
            vo.setRuleType(hit.getRuleType());
            vo.setPriority(hit.getPriority());
            vo.setReason(hit.getReason());
            vo.setAction(hit.getAction());
            vo.setConditionSource(hit.getConditionSource());
            return vo;
        }).toList();
    }

    private List<ChatVO.PromptRenderVO> toPromptRenderVOs(List<PromptRenderResult> renderResults) {
        if (renderResults == null) return List.of();
        return renderResults.stream().map(result -> {
            ChatVO.PromptRenderVO vo = new ChatVO.PromptRenderVO();
            vo.setPromptCode(result.getPromptCode());
            vo.setPromptName(result.getPromptName());
            vo.setPromptType(result.getPromptType());
            vo.setSceneType(result.getSceneType());
            vo.setRenderedContent(result.getRenderedContent());
            return vo;
        }).toList();
    }

    private ChatVO.IntentVO toIntentVO(AgentReply agentReply) {
        ChatVO.IntentVO intentVO = new ChatVO.IntentVO();
        intentVO.setIntentCode(agentReply.getIntentAnalysis().getIntentCode());
        intentVO.setIntentName(agentReply.getIntentAnalysis().getIntentName());
        intentVO.setConfidence(agentReply.getIntentAnalysis().getConfidence());
        return intentVO;
    }

    private ChatVO.KnowledgeRecallVO toKnowledgeRecallVO(KnowledgeRecall recall) {
        ChatVO.KnowledgeRecallVO vo = new ChatVO.KnowledgeRecallVO();
        vo.setProducts(new ArrayList<>());
        vo.setFaqs(new ArrayList<>());
        vo.setActivities(new ArrayList<>());

        for (KnowledgeEvidence evidence : recall.getEvidences()) {
            if (KnowledgeSource.PRODUCT == evidence.getSource()) {
                vo.getProducts().add(toProductRecallVO(evidence));
            } else if (KnowledgeSource.ACTIVITY == evidence.getSource()) {
                vo.getActivities().add(toActivityRecallVO(evidence));
            } else if (KnowledgeSource.FAQ == evidence.getSource()) {
                vo.getFaqs().add(toFaqRecallVO(evidence));
            }
        }
        return vo;
    }

    private ChatVO.ProductRecallVO toProductRecallVO(KnowledgeEvidence evidence) {
        ChatVO.ProductRecallVO vo = new ChatVO.ProductRecallVO();
        vo.setProductId(evidence.getEvidenceId());
        vo.setProductName(evidence.getTitle());
        vo.setScore(evidence.getScore());
        vo.setSource(String.valueOf(attributeOrDefault(evidence, "matchMode", evidence.getSource().name())));
        return vo;
    }

    private ChatVO.ActivityRecallVO toActivityRecallVO(KnowledgeEvidence evidence) {
        ChatVO.ActivityRecallVO vo = new ChatVO.ActivityRecallVO();
        vo.setActivityId(evidence.getEvidenceId());
        vo.setActivityName(evidence.getTitle());
        vo.setDescription(evidence.getContent());
        vo.setScore(evidence.getScore());
        return vo;
    }

    private ChatVO.FaqRecallVO toFaqRecallVO(KnowledgeEvidence evidence) {
        ChatVO.FaqRecallVO vo = new ChatVO.FaqRecallVO();
        vo.setFaqId(evidence.getEvidenceId());
        vo.setQuestion(evidence.getTitle());
        vo.setAnswer(evidence.getContent());
        vo.setScore(evidence.getScore());
        return vo;
    }

    private List<ChatVO.ReasoningStepVO> toReasoningProcess(List<ReasoningStep> reasoningSteps) {
        List<ChatVO.ReasoningStepVO> result = new ArrayList<>();
        if (reasoningSteps == null) return result;

        for (ReasoningStep step : reasoningSteps) {
            ChatVO.ReasoningStepVO vo = new ChatVO.ReasoningStepVO();
            vo.setStep(step.getStep());
            vo.setTitle(step.getTitle());
            vo.setContent(step.getContent());
            vo.setTimestamp(step.getTimestamp());
            result.add(vo);
        }
        return result;
    }

    private ChatVO.MessageVO toMessageVO(ChatMessage message) {
        ChatVO.MessageVO vo = new ChatVO.MessageVO();
        vo.setMessageId(message.getMessageId());
        vo.setSessionId(message.getSessionId());
        vo.setRole(message.getRole());
        vo.setContent(message.getContent());
        vo.setCardType(message.getCardType());
        vo.setCreatedAt(message.getCreatedAt());
        return vo;
    }

    private List<ChatVO.RecentSessionVO> loadRecentSessions() {
        return sessionRepository.findAllByOrderByCreatedAtDesc().stream()
                .limit(5)
                .map(this::toRecentSessionVO)
                .toList();
    }

    private ChatVO.RecentSessionVO toRecentSessionVO(ChatSession session) {
        ChatVO.RecentSessionVO vo = new ChatVO.RecentSessionVO();
        vo.setSessionId(session.getSessionId());
        vo.setUserName(session.getUserName());
        vo.setChannel(session.getChannel());
        vo.setStatus(session.getStatus());
        vo.setUpdatedAt(session.getUpdatedAt());
        vo.setMessageCount((int) messageRepository.countBySessionId(session.getSessionId()));

        List<ChatMessage> messages = messageRepository.findBySessionIdOrderByCreatedAtAsc(session.getSessionId());
        if (!messages.isEmpty()) {
            vo.setLastMessage(messages.get(messages.size() - 1).getContent());
        }
        return vo;
    }

    private void saveAgentAudit(String sessionId, ChatMessage aiMessage, AgentReply agentReply) {
        ChatAgentAudit audit = new ChatAgentAudit();
        audit.setAuditId(UUID.randomUUID().toString());
        audit.setSessionId(sessionId);
        audit.setMessageId(aiMessage.getMessageId());
        audit.setReplyEngine(agentReply.getEngine() != null ? agentReply.getEngine().name() : null);

        if (agentReply.getIntentAnalysis() != null) {
            audit.setSceneType(agentReply.getIntentAnalysis().getSceneType());
            audit.setIntentCode(agentReply.getIntentAnalysis().getIntentCode());
            audit.setIntentName(agentReply.getIntentAnalysis().getIntentName());
            audit.setConfidence(agentReply.getIntentAnalysis().getConfidence());
        }
        if (agentReply.getKnowledgeRecall() != null && agentReply.getKnowledgeRecall().getEvidences() != null) {
            audit.setKnowledgeEvidenceCount(agentReply.getKnowledgeRecall().getEvidences().size());
        } else {
            audit.setKnowledgeEvidenceCount(0);
        }
        if (agentReply.getGuardrailDecision() != null) {
            audit.setSafe(agentReply.getGuardrailDecision().isSafe());
            audit.setFallbackRequired(agentReply.getGuardrailDecision().isFallbackRequired());
            audit.setFallbackReason(agentReply.getGuardrailDecision().getFallbackReason() != null
                    ? agentReply.getGuardrailDecision().getFallbackReason().name()
                    : null);
            audit.setTransferRecommended(agentReply.getGuardrailDecision().isTransferRecommended());
            audit.setTransferReason(agentReply.getGuardrailDecision().getTransferReason());
            audit.setTransferPriority(agentReply.getGuardrailDecision().getTransferPriority());
            audit.setRuleHitCount(agentReply.getGuardrailDecision().getRuleHits() != null
                    ? agentReply.getGuardrailDecision().getRuleHits().size()
                    : 0);
            audit.setRuleHitCodes(joinCodes(agentReply.getGuardrailDecision().getRuleHits()));
        } else {
            audit.setSafe(true);
            audit.setFallbackRequired(false);
            audit.setTransferRecommended(false);
            audit.setRuleHitCount(0);
        }
        audit.setPromptRenderCount(agentReply.getPromptRenderResults() != null
                ? agentReply.getPromptRenderResults().size()
                : 0);
        audit.setPromptCodes(joinPromptCodes(agentReply.getPromptRenderResults()));
        auditRepository.save(audit);
    }

    private String joinCodes(List<RuleHit> ruleHits) {
        if (ruleHits == null) return null;
        String codes = ruleHits.stream()
                .map(RuleHit::getRuleCode)
                .filter(Objects::nonNull)
                .filter(code -> !code.isBlank())
                .distinct()
                .collect(Collectors.joining(","));
        return codes.isBlank() ? null : codes;
    }

    private String joinPromptCodes(List<PromptRenderResult> promptResults) {
        if (promptResults == null) return null;
        String codes = promptResults.stream()
                .map(PromptRenderResult::getPromptCode)
                .filter(Objects::nonNull)
                .filter(code -> !code.isBlank())
                .distinct()
                .collect(Collectors.joining(","));
        return codes.isBlank() ? null : codes;
    }

    private List<ChatVO.AgentAuditVO> loadRecentAudits() {
        return auditRepository.findTop5ByOrderByCreatedAtDesc().stream()
                .map(this::toAgentAuditVO)
                .toList();
    }

    private List<ChatVO.AgentAuditVO> loadSessionAudits(String sessionId) {
        return auditRepository.findBySessionIdOrderByCreatedAtAsc(sessionId).stream()
                .map(this::toAgentAuditVO)
                .toList();
    }

    private ChatVO.AgentAuditVO toAgentAuditVO(ChatAgentAudit audit) {
        ChatVO.AgentAuditVO vo = new ChatVO.AgentAuditVO();
        vo.setSessionId(audit.getSessionId());
        vo.setMessageId(audit.getMessageId());
        vo.setSceneType(audit.getSceneType());
        vo.setIntentName(audit.getIntentName());
        vo.setConfidence(audit.getConfidence());
        vo.setReplyEngine(audit.getReplyEngine());
        vo.setSafe(audit.getSafe());
        vo.setFallbackRequired(audit.getFallbackRequired());
        vo.setFallbackReason(audit.getFallbackReason());
        vo.setTransferRecommended(audit.getTransferRecommended());
        vo.setTransferReason(audit.getTransferReason());
        vo.setTransferPriority(audit.getTransferPriority());
        vo.setKnowledgeEvidenceCount(audit.getKnowledgeEvidenceCount());
        vo.setRuleHitCount(audit.getRuleHitCount());
        vo.setPromptRenderCount(audit.getPromptRenderCount());
        vo.setRuleHitCodes(audit.getRuleHitCodes());
        vo.setPromptCodes(audit.getPromptCodes());
        vo.setCreatedAt(audit.getCreatedAt());
        return vo;
    }

    private ChatVO.TransferTicketVO ensureTransferTicket(String sessionId, ChatMessage aiMessage, AgentReply agentReply) {
        if (agentReply.getGuardrailDecision() == null || !agentReply.getGuardrailDecision().isTransferRecommended()) {
            return loadLatestTransferTicket(sessionId);
        }

        ChatTransferTicket ticket = transferTicketRepository
                .findFirstBySessionIdAndStatusOrderByCreatedAtDesc(sessionId, "PENDING")
                .orElseGet(() -> {
                    ChatTransferTicket created = new ChatTransferTicket();
                    created.setTicketId(UUID.randomUUID().toString());
                    created.setSessionId(sessionId);
                    created.setStatus("PENDING");
                    return created;
                });

        ticket.setMessageId(aiMessage.getMessageId());
        ticket.setPriority(agentReply.getGuardrailDecision().getTransferPriority());
        ticket.setReason(agentReply.getGuardrailDecision().getTransferReason());
        return toTransferTicketVO(transferTicketRepository.save(ticket));
    }

    private ChatVO.TransferTicketVO loadLatestTransferTicket(String sessionId) {
        return transferTicketRepository.findBySessionIdOrderByCreatedAtDesc(sessionId).stream()
                .findFirst()
                .map(this::toTransferTicketVO)
                .orElse(null);
    }

    private List<ChatVO.TransferTicketVO> listTransferTicketsBySession(String sessionId) {
        return transferTicketRepository.findBySessionIdOrderByCreatedAtDesc(sessionId).stream()
                .map(this::toTransferTicketVO)
                .toList();
    }

    private ChatVO.TransferTicketVO toTransferTicketVO(ChatTransferTicket ticket) {
        ChatVO.TransferTicketVO vo = new ChatVO.TransferTicketVO();
        vo.setTicketId(ticket.getTicketId());
        vo.setSessionId(ticket.getSessionId());
        vo.setMessageId(ticket.getMessageId());
        vo.setStatus(ticket.getStatus());
        vo.setPriority(ticket.getPriority());
        vo.setReason(ticket.getReason());
        vo.setAssignedAgentId(ticket.getAssignedAgentId());
        vo.setAssignedAgentName(ticket.getAssignedAgentName());
        vo.setResolutionNote(ticket.getResolutionNote());
        vo.setCreatedAt(ticket.getCreatedAt());
        vo.setUpdatedAt(ticket.getUpdatedAt());
        vo.setResolvedAt(ticket.getResolvedAt());
        return vo;
    }

    private ChatVO.TransferSummaryVO buildTransferSummary(LocalDateTime startOfDay, LocalDateTime now) {
        List<ChatTransferTicket> resolvedTickets = transferTicketRepository.findAll().stream()
                .filter(ticket -> "RESOLVED".equals(ticket.getStatus()))
                .filter(ticket -> ticket.getCreatedAt() != null && ticket.getResolvedAt() != null)
                .toList();

        ChatVO.TransferSummaryVO vo = new ChatVO.TransferSummaryVO();
        vo.setPendingTickets(transferTicketRepository.countByStatus("PENDING"));
        vo.setTodayCreatedTickets(transferTicketRepository.countByCreatedAtBetween(startOfDay, now));
        vo.setTodayResolvedTickets(transferTicketRepository.countByStatusAndResolvedAtBetween("RESOLVED", startOfDay, now));
        vo.setHighPriorityPendingTickets(transferTicketRepository.countByStatusAndPriority("PENDING", "HIGH"));
        vo.setAverageResolveMinutes(roundAverage(resolvedTickets.stream()
                .map(ticket -> (double) Duration.between(ticket.getCreatedAt(), ticket.getResolvedAt()).toMinutes())
                .toList()));
        return vo;
    }

    private List<ChatVO.TransferTicketVO> loadRecentTransferTickets() {
        return transferTicketRepository.findTop5ByOrderByCreatedAtDesc().stream()
                .map(this::toTransferTicketVO)
                .toList();
    }

    private ChatVO.QualitySummaryVO buildQualitySummary() {
        List<ChatAgentAudit> audits = auditRepository.findAll();
        ChatVO.QualitySummaryVO vo = new ChatVO.QualitySummaryVO();
        vo.setTotalAuditedReplies((long) audits.size());
        vo.setAverageConfidence(roundAverage(audits.stream()
                .map(ChatAgentAudit::getConfidence)
                .filter(value -> value != null)
                .toList()));
        vo.setFallbackRate(roundRate(countFallback(audits), audits.size()));
        vo.setUnsafeRate(roundRate(countUnsafe(audits), audits.size()));
        vo.setAverageKnowledgeEvidenceCount(roundAverage(audits.stream()
                .map(ChatAgentAudit::getKnowledgeEvidenceCount)
                .filter(value -> value != null)
                .map(Integer::doubleValue)
                .toList()));
        vo.setAverageRuleHitCount(roundAverage(audits.stream()
                .map(ChatAgentAudit::getRuleHitCount)
                .filter(value -> value != null)
                .map(Integer::doubleValue)
                .toList()));
        vo.setAveragePromptRenderCount(roundAverage(audits.stream()
                .map(ChatAgentAudit::getPromptRenderCount)
                .filter(value -> value != null)
                .map(Integer::doubleValue)
                .toList()));
        return vo;
    }

    private ChatVO.SessionQualityVO buildSessionQuality(String sessionId) {
        List<ChatAgentAudit> audits = auditRepository.findBySessionIdOrderByCreatedAtAsc(sessionId);
        long total = audits.size();
        long fallbackCount = countFallback(audits);
        long unsafeCount = countUnsafe(audits);
        long lowConfidenceCount = audits.stream()
                .filter(audit -> audit.getConfidence() != null && audit.getConfidence() < 0.6)
                .count();

        double averageConfidence = roundAverage(audits.stream()
                .map(ChatAgentAudit::getConfidence)
                .filter(value -> value != null)
                .toList());
        double fallbackRate = roundRate(fallbackCount, total);
        double unsafeRate = roundRate(unsafeCount, total);
        double lowConfidenceRate = roundRate(lowConfidenceCount, total);

        ChatVO.SessionQualityVO vo = new ChatVO.SessionQualityVO();
        vo.setTotalAuditedReplies(total);
        vo.setFallbackReplies(fallbackCount);
        vo.setUnsafeReplies(unsafeCount);
        vo.setLowConfidenceReplies(lowConfidenceCount);
        vo.setAverageConfidence(averageConfidence);
        vo.setFallbackRate(fallbackRate);
        vo.setUnsafeRate(unsafeRate);
        vo.setReliabilityScore(calculateReliabilityScore(fallbackRate, unsafeRate, lowConfidenceRate));
        vo.setRiskLevel(resolveRiskLevel(vo.getReliabilityScore(), unsafeCount, fallbackRate));
        vo.setPrimaryFallbackReason(resolvePrimaryFallbackReason(audits));
        return vo;
    }

    private double calculateReliabilityScore(double fallbackRate, double unsafeRate, double lowConfidenceRate) {
        double score = 100D - fallbackRate * 0.35D - unsafeRate * 0.45D - lowConfidenceRate * 0.20D;
        return Math.max(0D, Math.round(score * 100) / 100D);
    }

    private String resolveRiskLevel(double reliabilityScore, long unsafeCount, double fallbackRate) {
        if (unsafeCount > 0 || reliabilityScore < 70D || fallbackRate >= 50D) return "HIGH";
        if (reliabilityScore < 85D || fallbackRate > 0D) return "MEDIUM";
        return "LOW";
    }

    private String resolvePrimaryFallbackReason(List<ChatAgentAudit> audits) {
        return audits.stream()
                .map(ChatAgentAudit::getFallbackReason)
                .filter(reason -> reason != null && !reason.isBlank())
                .collect(Collectors.groupingBy(reason -> reason, Collectors.counting()))
                .entrySet()
                .stream()
                .max(Map.Entry.comparingByValue())
                .map(Map.Entry::getKey)
                .orElse(null);
    }

    private List<ChatVO.RuntimeTrendVO> buildDailyTrends() {
        LocalDate startDate = LocalDate.now().minusDays(6);
        LocalDateTime start = startDate.atStartOfDay();
        LocalDateTime end = LocalDate.now().plusDays(1).atStartOfDay();
        List<ChatAgentAudit> audits = auditRepository.findByCreatedAtBetweenOrderByCreatedAtAsc(start, end);
        Map<LocalDate, List<ChatAgentAudit>> auditsByDate = audits.stream()
                .filter(audit -> audit.getCreatedAt() != null)
                .collect(Collectors.groupingBy(audit -> audit.getCreatedAt().toLocalDate()));

        List<ChatVO.RuntimeTrendVO> trends = new ArrayList<>();
        for (int i = 0; i < 7; i++) {
            LocalDate date = startDate.plusDays(i);
            List<ChatAgentAudit> dayAudits = auditsByDate.getOrDefault(date, List.of());
            ChatVO.RuntimeTrendVO trend = new ChatVO.RuntimeTrendVO();
            trend.setDate(date.toString());
            trend.setReplies((long) dayAudits.size());
            trend.setFallbackReplies(countFallback(dayAudits));
            trend.setUnsafeReplies(countUnsafe(dayAudits));
            trend.setAverageConfidence(roundAverage(dayAudits.stream()
                    .map(ChatAgentAudit::getConfidence)
                    .filter(value -> value != null)
                    .toList()));
            trends.add(trend);
        }
        return trends;
    }

    private ChatVO.RuntimeSnapshotVO loadLatestSnapshot() {
        return snapshotRepository.findTop7ByOrderByCreatedAtDesc().stream()
                .findFirst()
                .map(this::toRuntimeSnapshotVO)
                .orElse(null);
    }

    private ChatVO.RuntimeSnapshotVO toRuntimeSnapshotVO(ChatRuntimeSnapshot snapshot) {
        ChatVO.RuntimeSnapshotVO vo = new ChatVO.RuntimeSnapshotVO();
        vo.setSnapshotId(snapshot.getSnapshotId());
        vo.setSnapshotDate(snapshot.getSnapshotDate() != null ? snapshot.getSnapshotDate().toString() : null);
        vo.setSnapshotType(snapshot.getSnapshotType());
        vo.setTotalSessions(snapshot.getTotalSessions());
        vo.setActiveSessions(snapshot.getActiveSessions());
        vo.setTotalMessages(snapshot.getTotalMessages());
        vo.setTotalAuditedReplies(snapshot.getTotalAuditedReplies());
        vo.setAverageConfidence(snapshot.getAverageConfidence());
        vo.setFallbackRate(snapshot.getFallbackRate());
        vo.setUnsafeRate(snapshot.getUnsafeRate());
        vo.setAverageKnowledgeEvidenceCount(snapshot.getAverageKnowledgeEvidenceCount());
        vo.setAverageRuleHitCount(snapshot.getAverageRuleHitCount());
        vo.setAveragePromptRenderCount(snapshot.getAveragePromptRenderCount());
        vo.setCreatedAt(snapshot.getCreatedAt());
        return vo;
    }

    private long countFallback(List<ChatAgentAudit> audits) {
        return audits.stream().filter(audit -> Boolean.TRUE.equals(audit.getFallbackRequired())).count();
    }

    private long countUnsafe(List<ChatAgentAudit> audits) {
        return audits.stream().filter(audit -> Boolean.FALSE.equals(audit.getSafe())).count();
    }

    private double roundRate(long numerator, long denominator) {
        if (denominator == 0) return 0D;
        return Math.round((double) numerator * 10000 / denominator) / 100D;
    }

    private double roundAverage(List<Double> values) {
        if (values.isEmpty()) return 0D;
        double total = values.stream().mapToDouble(Double::doubleValue).sum();
        return Math.round(total * 100 / values.size()) / 100D;
    }

    private ConversationMessage toConversationMessage(ChatMessage message) {
        ConversationMessage conversationMessage = new ConversationMessage();
        conversationMessage.setMessageId(message.getMessageId());
        conversationMessage.setRole("user".equals(message.getRole()) ? ConversationRole.USER : ConversationRole.ASSISTANT);
        conversationMessage.setContent(message.getContent());
        conversationMessage.setCreatedAt(message.getCreatedAt());
        return conversationMessage;
    }

    private Object attributeOrDefault(KnowledgeEvidence evidence, String key, Object defaultValue) {
        if (evidence.getAttributes() == null) return defaultValue;
        return evidence.getAttributes().getOrDefault(key, defaultValue);
    }
}
