package com.anjing.module.agent.runtime;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.RuleHit;
import com.anjing.module.scene.entity.Rule;
import com.anjing.module.scene.repository.RuleRepository;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * 轻量规则执行器。
 *
 * 支持内置规则码和轻量 JSON 条件表达式，让规则启用、命中和动作可解释。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class RuleEngine {

    private static final double TRANSFER_CONFIDENCE_THRESHOLD = 0.6;
    private static final List<String> SENSITIVE_WORDS = List.of("骗子", "骗钱", "欺诈", "威胁", "恐吓");

    private final RuleRepository ruleRepository;
    private final ObjectMapper objectMapper;

    public List<RuleHit> evaluate(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        List<RuleHit> hits = new ArrayList<>();
        for (Rule rule : ruleRepository.findByEnabledTrueOrderByPriorityAsc()) {
            RuleHit hit = evaluateRule(rule, turn, analysis, recall);
            if (hit != null) {
                hits.add(hit);
                rule.setTriggerCount(rule.getTriggerCount() == null ? 1 : rule.getTriggerCount() + 1);
                ruleRepository.save(rule);
            }
        }
        return hits;
    }

    private RuleHit evaluateRule(Rule rule, ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        Optional<RuleHit> expressionHit = evaluateJsonCondition(rule, turn, analysis, recall);
        if (expressionHit.isPresent()) {
            return expressionHit.get();
        }

        return switch (rule.getRuleCode()) {
            case "SENSITIVE_FILTER" -> matchSensitiveFilter(rule, turn);
            case "TRANSFER_THRESHOLD" -> matchTransferThreshold(rule, analysis);
            case "VIP_PRIORITY" -> matchVipPriority(rule, turn);
            default -> null;
        };
    }

    private Optional<RuleHit> evaluateJsonCondition(
            Rule rule,
            ConversationTurn turn,
            IntentAnalysis analysis,
            KnowledgeRecall recall
    ) {
        if (rule.getConditions() == null || rule.getConditions().isBlank()) {
            return Optional.empty();
        }

        try {
            JsonNode condition = objectMapper.readTree(rule.getConditions());
            if (!matches(condition, turn, analysis, recall)) {
                return Optional.empty();
            }
            RuleAction action = parseAction(rule);
            return Optional.of(toHit(rule, action.reason(), action.action(), "JSON_CONDITION"));
        } catch (Exception ex) {
            log.warn("规则条件解析失败: ruleCode={}, message={}", rule.getRuleCode(), ex.getMessage());
            return Optional.empty();
        }
    }

    private boolean matches(JsonNode node, ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        if (node == null || node.isNull()) return false;
        if (node.has("all")) {
            JsonNode items = node.get("all");
            if (!items.isArray() || items.isEmpty()) return false;
            for (JsonNode item : items) {
                if (!matches(item, turn, analysis, recall)) return false;
            }
            return true;
        }
        if (node.has("any")) {
            JsonNode items = node.get("any");
            if (!items.isArray() || items.isEmpty()) return false;
            for (JsonNode item : items) {
                if (matches(item, turn, analysis, recall)) return true;
            }
            return false;
        }
        return matchesSingle(node, turn, analysis, recall);
    }

    private boolean matchesSingle(JsonNode node, ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        String field = text(node, "field");
        String op = text(node, "op");
        JsonNode expected = node.get("value");
        Object actual = resolveField(field, turn, analysis, recall);

        return switch (op == null ? "eq" : op) {
            case "eq" -> equalsValue(actual, expected);
            case "ne" -> !equalsValue(actual, expected);
            case "contains" -> actual != null && expected != null && String.valueOf(actual).contains(expected.asText());
            case "not_contains" -> actual == null || expected == null || !String.valueOf(actual).contains(expected.asText());
            case "gt" -> canCompareNumber(actual, expected) && compareNumber(actual, expected) > 0;
            case "gte" -> canCompareNumber(actual, expected) && compareNumber(actual, expected) >= 0;
            case "lt" -> canCompareNumber(actual, expected) && compareNumber(actual, expected) < 0;
            case "lte" -> canCompareNumber(actual, expected) && compareNumber(actual, expected) <= 0;
            case "in" -> inValues(actual, expected);
            case "is_empty" -> actual == null || String.valueOf(actual).isBlank();
            case "is_not_empty" -> actual != null && !String.valueOf(actual).isBlank();
            default -> false;
        };
    }

    private Object resolveField(String field, ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        if (field == null || field.isBlank()) return null;
        return switch (field) {
            case "userMessage" -> turn.getUserMessage();
            case "sceneType" -> analysis.getSceneType();
            case "intentCode" -> analysis.getIntentCode();
            case "intentName" -> analysis.getIntentName();
            case "confidence" -> analysis.getConfidence();
            case "emotion" -> analysis.getEmotion();
            case "knowledgeCount" -> recall.getEvidences().size();
            case "hasReliableKnowledge" -> recall.hasReliableEvidence();
            default -> resolveContextField(field, turn);
        };
    }

    private Object resolveContextField(String field, ConversationTurn turn) {
        if (!field.startsWith("context.") || turn.getContext() == null) return null;
        Object current = turn.getContext();
        String[] parts = field.substring("context.".length()).split("\\.");
        for (String part : parts) {
            if (current instanceof Map<?, ?> map) {
                current = map.get(part);
            } else {
                return null;
            }
        }
        return current;
    }

    private RuleAction parseAction(Rule rule) {
        if (rule.getActions() == null || rule.getActions().isBlank()) {
            return new RuleAction(defaultReason(rule), defaultAction(rule));
        }

        try {
            JsonNode actionNode = objectMapper.readTree(rule.getActions());
            String reason = text(actionNode, "reason");
            String action = text(actionNode, "action");
            return new RuleAction(
                    reason == null || reason.isBlank() ? defaultReason(rule) : reason,
                    action == null || action.isBlank() ? defaultAction(rule) : action
            );
        } catch (Exception ex) {
            log.warn("规则动作解析失败: ruleCode={}, message={}", rule.getRuleCode(), ex.getMessage());
            return new RuleAction(defaultReason(rule), defaultAction(rule));
        }
    }

    private RuleHit matchSensitiveFilter(Rule rule, ConversationTurn turn) {
        if (turn.getUserMessage() == null) return null;
        boolean hit = SENSITIVE_WORDS.stream().anyMatch(turn.getUserMessage()::contains);
        if (!hit) return null;
        return toHit(rule, "用户输入命中高风险表达", "SAFETY_BLOCK", "BUILT_IN");
    }

    private RuleHit matchTransferThreshold(Rule rule, IntentAnalysis analysis) {
        if (analysis.getConfidence() == null || analysis.getConfidence() >= TRANSFER_CONFIDENCE_THRESHOLD) {
            return null;
        }
        return toHit(rule, "意图置信度低于 " + TRANSFER_CONFIDENCE_THRESHOLD, "TRANSFER_OR_CLARIFY", "BUILT_IN");
    }

    private RuleHit matchVipPriority(Rule rule, ConversationTurn turn) {
        Object userProfile = turn.getContext() != null ? turn.getContext().get("userProfile") : null;
        if (userProfile == null) return null;
        String profileText = String.valueOf(userProfile);
        if (!profileText.contains("VIP")) return null;
        return toHit(rule, "用户画像命中 VIP", "PRIORITY_RESPONSE", "BUILT_IN");
    }

    private RuleHit toHit(Rule rule, String reason, String action, String conditionSource) {
        RuleHit hit = new RuleHit();
        hit.setRuleCode(rule.getRuleCode());
        hit.setRuleName(rule.getRuleName());
        hit.setRuleType(rule.getRuleType());
        hit.setPriority(rule.getPriority());
        hit.setReason(reason);
        hit.setAction(action);
        hit.setConditionSource(conditionSource);
        return hit;
    }

    private boolean equalsValue(Object actual, JsonNode expected) {
        if (actual == null || expected == null || expected.isNull()) {
            return actual == null && (expected == null || expected.isNull());
        }
        if (actual instanceof Number || expected.isNumber()) {
            if (!expected.isNumber() || Double.isNaN(toDouble(actual))) return false;
            return Double.compare(toDouble(actual), expected.asDouble()) == 0;
        }
        if (actual instanceof Boolean || expected.isBoolean()) {
            return Boolean.parseBoolean(String.valueOf(actual)) == expected.asBoolean();
        }
        return String.valueOf(actual).equals(expected.asText());
    }

    private int compareNumber(Object actual, JsonNode expected) {
        return Double.compare(toDouble(actual), expected.asDouble());
    }

    private boolean canCompareNumber(Object actual, JsonNode expected) {
        return actual != null && expected != null && expected.isNumber() && !Double.isNaN(toDouble(actual));
    }

    private boolean inValues(Object actual, JsonNode expected) {
        if (actual == null || expected == null || !expected.isArray()) return false;
        for (JsonNode value : expected) {
            if (equalsValue(actual, value)) return true;
        }
        return false;
    }

    private double toDouble(Object value) {
        if (value instanceof Number number) return number.doubleValue();
        try {
            return Double.parseDouble(String.valueOf(value));
        } catch (NumberFormatException ex) {
            return Double.NaN;
        }
    }

    private String text(JsonNode node, String field) {
        JsonNode value = node != null ? node.get(field) : null;
        return value != null && !value.isNull() ? value.asText() : null;
    }

    private String defaultReason(Rule rule) {
        return rule.getDescription() != null && !rule.getDescription().isBlank()
                ? rule.getDescription()
                : "命中规则条件";
    }

    private String defaultAction(Rule rule) {
        return rule.getRuleType() != null && !rule.getRuleType().isBlank()
                ? rule.getRuleType()
                : "OBSERVE";
    }

    private record RuleAction(String reason, String action) {
    }
}
