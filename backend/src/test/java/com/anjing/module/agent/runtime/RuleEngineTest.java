package com.anjing.module.agent.runtime;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.RuleHit;
import com.anjing.module.scene.entity.Rule;
import com.anjing.module.scene.repository.RuleRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

class RuleEngineTest {

    private final RuleRepository ruleRepository = mock(RuleRepository.class);
    private final RuleEngine ruleEngine = new RuleEngine(ruleRepository, new ObjectMapper());

    @Test
    void evaluatesJsonConditionAndAction() {
        Rule rule = new Rule();
        rule.setRuleCode("LOW_RETURN_CONFIDENCE");
        rule.setRuleName("低置信度退换货");
        rule.setRuleType("转接规则");
        rule.setPriority(1);
        rule.setTriggerCount(0);
        rule.setConditions("""
                {
                  "all": [
                    { "field": "intentCode", "op": "eq", "value": "RETURN_EXCHANGE" },
                    { "field": "confidence", "op": "lt", "value": 0.7 }
                  ]
                }
                """);
        rule.setActions("""
                {
                  "reason": "退换货意图置信度不足，需要澄清",
                  "action": "TRANSFER_OR_CLARIFY"
                }
                """);
        when(ruleRepository.findByEnabledTrueOrderByPriorityAsc()).thenReturn(List.of(rule));

        List<RuleHit> hits = ruleEngine.evaluate(turn("我想退货"), analysis("RETURN_EXCHANGE", 0.62), new KnowledgeRecall());

        assertThat(hits).hasSize(1);
        RuleHit hit = hits.get(0);
        assertThat(hit.getRuleCode()).isEqualTo("LOW_RETURN_CONFIDENCE");
        assertThat(hit.getAction()).isEqualTo("TRANSFER_OR_CLARIFY");
        assertThat(hit.getReason()).isEqualTo("退换货意图置信度不足，需要澄清");
        assertThat(hit.getConditionSource()).isEqualTo("JSON_CONDITION");
        assertThat(rule.getTriggerCount()).isEqualTo(1);
        verify(ruleRepository).save(rule);
    }

    @Test
    void keepsBuiltInRulesWhenConditionIsBlank() {
        Rule rule = new Rule();
        rule.setRuleCode("SENSITIVE_FILTER");
        rule.setRuleName("敏感词过滤");
        rule.setRuleType("安全规则");
        rule.setPriority(1);
        rule.setTriggerCount(0);
        when(ruleRepository.findByEnabledTrueOrderByPriorityAsc()).thenReturn(List.of(rule));

        List<RuleHit> hits = ruleEngine.evaluate(turn("你们是不是骗钱"), analysis("GENERAL", 0.9), new KnowledgeRecall());

        assertThat(hits).hasSize(1);
        assertThat(hits.get(0).getAction()).isEqualTo("SAFETY_BLOCK");
        assertThat(hits.get(0).getConditionSource()).isEqualTo("BUILT_IN");
        assertThat(rule.getTriggerCount()).isEqualTo(1);
        verify(ruleRepository).save(rule);
    }

    private ConversationTurn turn(String message) {
        ConversationTurn turn = new ConversationTurn();
        turn.setUserMessage(message);
        return turn;
    }

    private IntentAnalysis analysis(String intentCode, double confidence) {
        IntentAnalysis analysis = new IntentAnalysis();
        analysis.setSceneType("售后咨询");
        analysis.setIntentCode(intentCode);
        analysis.setIntentName(intentCode);
        analysis.setConfidence(confidence);
        analysis.setEmotion("中性");
        return analysis;
    }
}
