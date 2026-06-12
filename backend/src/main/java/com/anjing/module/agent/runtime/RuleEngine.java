package com.anjing.module.agent.runtime;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.RuleHit;
import com.anjing.module.scene.entity.Rule;
import com.anjing.module.scene.repository.RuleRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

/**
 * 轻量规则执行器。
 *
 * 当前先支持内置规则码，让规则启用、命中和动作可解释；后续再扩展 JSON 条件表达式。
 */
@Service
@RequiredArgsConstructor
public class RuleEngine {

    private static final double TRANSFER_CONFIDENCE_THRESHOLD = 0.6;
    private static final List<String> SENSITIVE_WORDS = List.of("骗子", "骗钱", "欺诈", "威胁", "恐吓");

    private final RuleRepository ruleRepository;

    public List<RuleHit> evaluate(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        List<RuleHit> hits = new ArrayList<>();
        for (Rule rule : ruleRepository.findByEnabledTrueOrderByPriorityAsc()) {
            RuleHit hit = evaluateRule(rule, turn, analysis, recall);
            if (hit != null) {
                hits.add(hit);
            }
        }
        return hits;
    }

    private RuleHit evaluateRule(Rule rule, ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        return switch (rule.getRuleCode()) {
            case "SENSITIVE_FILTER" -> matchSensitiveFilter(rule, turn);
            case "TRANSFER_THRESHOLD" -> matchTransferThreshold(rule, analysis);
            case "VIP_PRIORITY" -> matchVipPriority(rule, turn);
            default -> null;
        };
    }

    private RuleHit matchSensitiveFilter(Rule rule, ConversationTurn turn) {
        if (turn.getUserMessage() == null) return null;
        boolean hit = SENSITIVE_WORDS.stream().anyMatch(turn.getUserMessage()::contains);
        if (!hit) return null;
        return toHit(rule, "用户输入命中高风险表达", "SAFETY_BLOCK");
    }

    private RuleHit matchTransferThreshold(Rule rule, IntentAnalysis analysis) {
        if (analysis.getConfidence() == null || analysis.getConfidence() >= TRANSFER_CONFIDENCE_THRESHOLD) {
            return null;
        }
        return toHit(rule, "意图置信度低于 " + TRANSFER_CONFIDENCE_THRESHOLD, "TRANSFER_OR_CLARIFY");
    }

    private RuleHit matchVipPriority(Rule rule, ConversationTurn turn) {
        Object userProfile = turn.getContext() != null ? turn.getContext().get("userProfile") : null;
        if (userProfile == null) return null;
        String profileText = String.valueOf(userProfile);
        if (!profileText.contains("VIP")) return null;
        return toHit(rule, "用户画像命中 VIP", "PRIORITY_RESPONSE");
    }

    private RuleHit toHit(Rule rule, String reason, String action) {
        RuleHit hit = new RuleHit();
        hit.setRuleCode(rule.getRuleCode());
        hit.setRuleName(rule.getRuleName());
        hit.setRuleType(rule.getRuleType());
        hit.setPriority(rule.getPriority());
        hit.setReason(reason);
        hit.setAction(action);
        return hit;
    }
}
