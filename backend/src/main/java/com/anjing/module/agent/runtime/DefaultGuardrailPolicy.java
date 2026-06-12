package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.GuardrailPolicy;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.FallbackReason;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

/**
 * 当前版本的可靠性护栏策略。
 */
@Service
public class DefaultGuardrailPolicy implements GuardrailPolicy {

    private static final double LOW_CONFIDENCE_THRESHOLD = 0.6;

    @Override
    public GuardrailDecision decide(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        GuardrailDecision decision = new GuardrailDecision();
        decision.setSafe(true);
        decision.setFallbackRequired(false);
        decision.setPolicyTags(new ArrayList<>());

        List<String> tags = decision.getPolicyTags();
        if (analysis.getConfidence() != null && analysis.getConfidence() < LOW_CONFIDENCE_THRESHOLD) {
            decision.setFallbackRequired(true);
            decision.setFallbackReason(FallbackReason.LOW_CONFIDENCE);
            decision.setUserVisibleNotice("我还不太确定您的具体诉求，可以请您补充更多信息。");
            tags.add("low-confidence");
            return decision;
        }

        if (!recall.hasReliableEvidence() && requiresEvidence(analysis.getIntentCode())) {
            decision.setFallbackRequired(true);
            decision.setFallbackReason(FallbackReason.NO_RELIABLE_KNOWLEDGE);
            decision.setUserVisibleNotice("当前没有检索到足够可靠的知识，将使用标准客服规则回答。");
            tags.add("no-reliable-knowledge");
        }

        if (analysis.getEngine() != null) {
            tags.add("analysis-engine-" + analysis.getEngine().name().toLowerCase());
        }
        return decision;
    }

    private boolean requiresEvidence(String intentCode) {
        return "PRODUCT_DISCOUNT".equals(intentCode) || "SIZE_CONSULT".equals(intentCode);
    }
}
