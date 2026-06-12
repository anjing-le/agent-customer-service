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
import com.anjing.module.chat.entity.ChatMessage;
import com.anjing.module.chat.entity.ChatSession;
import com.anjing.module.chat.repository.ChatMessageRepository;
import com.anjing.module.chat.repository.ChatSessionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.PageRequest;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

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
        vo.setAverageMessagesPerSession(totalSessions == 0 ? 0D : Math.round((double) totalMessages * 100 / totalSessions) / 100D);
        vo.setRecentSessions(loadRecentSessions());
        return vo;
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
        return toSendMessageVO(agentReply, aiMessage);
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

    private ChatVO.SendMessageVO toSendMessageVO(AgentReply agentReply, ChatMessage aiMessage) {
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
