package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.GuardrailPolicy;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.FallbackReason;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.RuleHit;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

/**
 * 当前版本的可靠性护栏策略。
 */
@Service
@RequiredArgsConstructor
public class DefaultGuardrailPolicy implements GuardrailPolicy {

    private final RuleEngine ruleEngine;

    @Override
    public GuardrailDecision decide(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        GuardrailDecision decision = new GuardrailDecision();
        decision.setSafe(true);
        decision.setFallbackRequired(false);
        decision.setPolicyTags(new ArrayList<>());
        decision.setRuleHits(ruleEngine.evaluate(turn, analysis, recall));

        List<String> tags = decision.getPolicyTags();
        for (RuleHit hit : decision.getRuleHits()) {
            tags.add("rule-hit-" + hit.getRuleCode());
        }

        RuleHit safetyHit = findHit(decision.getRuleHits(), "SENSITIVE_FILTER");
        if (safetyHit != null) {
            decision.setSafe(false);
            decision.setFallbackRequired(true);
            decision.setFallbackReason(FallbackReason.SAFETY_BLOCKED);
            decision.setUserVisibleNotice("检测到高风险表达，将使用安全客服话术并建议人工跟进。");
            return decision;
        }

        RuleHit transferHit = findHit(decision.getRuleHits(), "TRANSFER_THRESHOLD");
        if (transferHit != null) {
            decision.setFallbackRequired(true);
            decision.setFallbackReason(FallbackReason.LOW_CONFIDENCE);
            decision.setUserVisibleNotice("我还不太确定您的具体诉求，可以请您补充更多信息。");
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

    private RuleHit findHit(List<RuleHit> hits, String ruleCode) {
        return hits.stream()
                .filter(hit -> ruleCode.equals(hit.getRuleCode()))
                .findFirst()
                .orElse(null);
    }

    private boolean requiresEvidence(String intentCode) {
        return "PRODUCT_DISCOUNT".equals(intentCode) || "SIZE_CONSULT".equals(intentCode);
    }
}
