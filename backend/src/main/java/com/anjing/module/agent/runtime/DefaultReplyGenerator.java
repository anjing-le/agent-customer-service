package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.ReplyGenerator;
import com.anjing.module.agent.domain.AgentEngine;
import com.anjing.module.agent.domain.AgentReply;
import com.anjing.module.agent.domain.ConversationMessage;
import com.anjing.module.agent.domain.ConversationRole;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeEvidence;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.KnowledgeSource;
import com.anjing.module.chat.LlmService;
import com.anjing.module.scene.entity.Prompt;
import com.anjing.module.scene.repository.PromptRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * LLM 优先、规则兜底的回复生成器。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class DefaultReplyGenerator implements ReplyGenerator {

    private final LlmService llmService;
    private final PromptRepository promptRepository;

    @Override
    public AgentReply generate(
            ConversationTurn turn,
            IntentAnalysis analysis,
            KnowledgeRecall recall,
            GuardrailDecision guardrailDecision
    ) {
        AgentReply reply = new AgentReply();
        reply.setIntentAnalysis(analysis);
        reply.setKnowledgeRecall(recall);
        reply.setGuardrailDecision(guardrailDecision);
        reply.setCardType(selectTool(analysis, recall));

        String llmReply = null;
        if (!guardrailDecision.isFallbackRequired()) {
            llmReply = llmService.generateReply(turn.getUserMessage(), buildLlmContext(analysis, recall), buildChatHistory(turn));
        }

        if (llmReply != null) {
            reply.setContent(llmReply);
            reply.setEngine(AgentEngine.LLM);
        } else {
            reply.setContent(generateRuleReply(analysis, recall));
            reply.setEngine(AgentEngine.RULE);
        }

        log.info("Agent 回复生成完成: engine={}, fallback={}, cardType={}",
                reply.getEngine(), guardrailDecision.isFallbackRequired(), reply.getCardType());
        return reply;
    }

    private Map<String, Object> buildLlmContext(IntentAnalysis analysis, KnowledgeRecall recall) {
        Map<String, Object> context = new HashMap<>();
        context.put("sceneType", analysis.getSceneType());
        context.put("intentName", analysis.getIntentName());
        context.put("emotion", analysis.getEmotion());
        context.put("knowledge", buildKnowledgeText(recall));
        context.put("runtimePrompt", buildRuntimePrompt(analysis));
        return context;
    }

    private String buildRuntimePrompt(IntentAnalysis analysis) {
        List<Prompt> enabledSystemPrompts = promptRepository.findByStatusAndPromptType("启用", "SYSTEM");
        StringBuilder builder = new StringBuilder();
        for (Prompt prompt : enabledSystemPrompts) {
            if (!matchesScene(prompt, analysis)) continue;
            if (builder.length() > 0) builder.append("\n");
            builder.append("- ").append(prompt.getPromptName()).append(": ").append(prompt.getPromptContent());
        }
        return builder.length() > 0 ? builder.toString() : null;
    }

    private boolean matchesScene(Prompt prompt, IntentAnalysis analysis) {
        return prompt.getSceneType() == null
                || prompt.getSceneType().isBlank()
                || prompt.getSceneType().equals(analysis.getSceneType());
    }

    private List<Map<String, String>> buildChatHistory(ConversationTurn turn) {
        if (turn.getRecentHistory() == null || turn.getRecentHistory().isEmpty()) {
            return List.of();
        }

        List<Map<String, String>> history = new ArrayList<>();
        for (ConversationMessage message : turn.getRecentHistory()) {
            String role = message.getRole() == ConversationRole.USER ? "user" : "assistant";
            history.add(Map.of("role", role, "content", message.getContent()));
        }
        return history;
    }

    private String buildKnowledgeText(KnowledgeRecall recall) {
        StringBuilder builder = new StringBuilder();
        appendEvidenceGroup(builder, recall, KnowledgeSource.PRODUCT, "相关商品");
        appendEvidenceGroup(builder, recall, KnowledgeSource.ACTIVITY, "优惠活动");
        appendEvidenceGroup(builder, recall, KnowledgeSource.FAQ, "FAQ参考");
        return builder.length() > 0 ? builder.toString() : null;
    }

    private void appendEvidenceGroup(StringBuilder builder, KnowledgeRecall recall, KnowledgeSource source, String title) {
        List<KnowledgeEvidence> evidences = recall.getEvidences().stream()
                .filter(evidence -> source == evidence.getSource())
                .toList();
        if (evidences.isEmpty()) return;

        builder.append("\n【").append(title).append("】");
        for (KnowledgeEvidence evidence : evidences) {
            builder.append("\n- ").append(evidence.getTitle());
            if (evidence.getContent() != null && !evidence.getContent().isBlank()) {
                builder.append("：").append(evidence.getContent());
            }
        }
    }

    private String generateRuleReply(IntentAnalysis analysis, KnowledgeRecall recall) {
        StringBuilder reply = new StringBuilder();

        firstEvidence(recall, KnowledgeSource.FAQ).ifPresent(evidence -> reply.append(evidence.getContent()));

        List<KnowledgeEvidence> activities = evidencesOf(recall, KnowledgeSource.ACTIVITY);
        if (!activities.isEmpty()) {
            if (reply.length() > 0) reply.append("\n\n");
            reply.append("当前可用优惠活动：");
            for (KnowledgeEvidence activity : activities) {
                reply.append("\n- ").append(activity.getTitle()).append("：").append(activity.getContent());
            }
        }

        List<KnowledgeEvidence> products = evidencesOf(recall, KnowledgeSource.PRODUCT);
        if (!products.isEmpty() && "SIZE_CONSULT".equals(analysis.getIntentCode())) {
            if (reply.length() > 0) reply.append("\n\n");
            reply.append("相关商品：");
            for (KnowledgeEvidence product : products) {
                reply.append("\n- ").append(product.getTitle());
            }
        }

        if (reply.length() == 0) {
            return switch (analysis.getIntentCode()) {
                case "RETURN_EXCHANGE" -> "退换货步骤：\n1. 进入【我的订单】找到对应订单\n2. 点击【申请售后】\n3. 选择退换货原因并提交\n\n审核通过后，退款将在3-5个工作日内到账。";
                case "LOGISTICS_QUERY" -> "您可以在订单详情页查看物流信息。一般发货后2-3天送达，如需修改地址请在发货前联系客服。";
                case "PRODUCT_DISCOUNT" -> "当前没有检索到明确可用的优惠活动。建议您查看商品详情页优惠信息，或联系人工客服确认最新活动。";
                default -> "感谢您的咨询！请问还有什么其他问题需要帮助吗？";
            };
        }

        return reply.toString();
    }

    private String selectTool(IntentAnalysis analysis, KnowledgeRecall recall) {
        return switch (analysis.getIntentCode()) {
            case "PRODUCT_DISCOUNT" -> "ACTIVITY_CARD";
            case "SIZE_CONSULT" -> "PRODUCT_CARD";
            case "LOGISTICS_QUERY" -> "LOGISTICS_CARD";
            default -> null;
        };
    }

    private java.util.Optional<KnowledgeEvidence> firstEvidence(KnowledgeRecall recall, KnowledgeSource source) {
        return recall.getEvidences().stream().filter(evidence -> source == evidence.getSource()).findFirst();
    }

    private List<KnowledgeEvidence> evidencesOf(KnowledgeRecall recall, KnowledgeSource source) {
        return recall.getEvidences().stream().filter(evidence -> source == evidence.getSource()).toList();
    }
}
