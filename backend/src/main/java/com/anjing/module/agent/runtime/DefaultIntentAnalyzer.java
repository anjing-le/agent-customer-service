package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.IntentAnalyzer;
import com.anjing.module.agent.domain.AgentEngine;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.chat.LlmService;
import com.anjing.module.scene.entity.Intent;
import com.anjing.module.scene.repository.IntentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * LLM 优先、关键词兜底的意图分析实现。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class DefaultIntentAnalyzer implements IntentAnalyzer {

    private final LlmService llmService;
    private final IntentRepository intentRepository;

    @Override
    public IntentAnalysis analyze(ConversationTurn turn) {
        Map<String, String> llmAnalysis = llmService.analyzeUserInput(turn.getUserMessage());
        if (llmAnalysis != null) {
            IntentAnalysis analysis = new IntentAnalysis();
            analysis.setSceneType(llmAnalysis.get("scene"));
            analysis.setIntentCode(llmAnalysis.get("intentCode"));
            analysis.setIntentName(llmAnalysis.get("intentName"));
            analysis.setConfidence(parseConfidence(llmAnalysis.get("confidence")));
            analysis.setEmotion(llmAnalysis.get("emotion"));
            analysis.setEngine(AgentEngine.LLM);
            log.info("Agent 分析完成: engine=LLM, scene={}, intent={}, confidence={}, emotion={}",
                    analysis.getSceneType(), analysis.getIntentName(), analysis.getConfidence(), analysis.getEmotion());
            return analysis;
        }

        Intent configuredIntent = matchConfiguredIntent(turn.getUserMessage());
        IntentAnalysis fallback = configuredIntent != null
                ? fromConfiguredIntent(configuredIntent)
                : fromBuiltinFallback(turn.getUserMessage());
        fallback.setEmotion(analyzeEmotionFallback(turn.getUserMessage()));
        fallback.setEngine(AgentEngine.RULE);
        log.info("Agent 分析完成: engine=RULE, scene={}, intent={}, confidence={}, emotion={}",
                fallback.getSceneType(), fallback.getIntentName(), fallback.getConfidence(), fallback.getEmotion());
        return fallback;
    }

    private Intent matchConfiguredIntent(String content) {
        List<Intent> enabledIntents = intentRepository.findByStatusOrderByPriorityAsc("启用");
        for (Intent intent : enabledIntents) {
            if (matchesKeywords(content, intent.getTriggerKeywords())) {
                log.info("命中 Scene Intent 配置: code={}, name={}", intent.getIntentCode(), intent.getIntentName());
                return intent;
            }
        }
        return null;
    }

    private IntentAnalysis fromConfiguredIntent(Intent intent) {
        IntentAnalysis analysis = new IntentAnalysis();
        analysis.setSceneType(intent.getSceneType() != null ? intent.getSceneType() : "通用咨询");
        analysis.setIntentCode(intent.getIntentCode() != null ? intent.getIntentCode() : "GENERAL_QUERY");
        analysis.setIntentName(intent.getIntentName() != null ? intent.getIntentName() : intentNameOf(analysis.getIntentCode()));
        analysis.setConfidence(intent.getConfidenceThreshold() != null ? intent.getConfidenceThreshold() : 0.8);
        return analysis;
    }

    private IntentAnalysis fromBuiltinFallback(String content) {
        IntentAnalysis analysis = new IntentAnalysis();
        analysis.setSceneType(recognizeSceneFallback(content));
        analysis.setIntentCode(recognizeIntentCodeFallback(content));
        analysis.setIntentName(intentNameOf(analysis.getIntentCode()));
        analysis.setConfidence(confidenceOf(analysis.getIntentCode()));
        return analysis;
    }

    private Double parseConfidence(String confidence) {
        try {
            return Double.parseDouble(confidence);
        } catch (Exception e) {
            return 0.85;
        }
    }

    private String recognizeSceneFallback(String content) {
        if (containsAny(content, "优惠", "活动", "打折", "促销", "价格", "多少钱")) return "售前咨询";
        if (containsAny(content, "退货", "换货", "退款", "售后", "投诉", "质量")) return "售后服务";
        if (containsAny(content, "快递", "物流", "发货", "到货", "配送")) return "物流配送";
        if (containsAny(content, "支付", "付款", "发票", "账单")) return "支付问题";
        return "通用咨询";
    }

    private String recognizeIntentCodeFallback(String content) {
        if (containsAny(content, "优惠", "活动", "打折", "促销")) return "PRODUCT_DISCOUNT";
        if (containsAny(content, "退货", "换货", "退款")) return "RETURN_EXCHANGE";
        if (containsAny(content, "尺码", "大小", "尺寸")) return "SIZE_CONSULT";
        if (containsAny(content, "快递", "物流", "发货")) return "LOGISTICS_QUERY";
        return "GENERAL_QUERY";
    }

    private String intentNameOf(String intentCode) {
        return switch (intentCode) {
            case "PRODUCT_DISCOUNT" -> "商品优惠查询";
            case "RETURN_EXCHANGE" -> "退换货咨询";
            case "SIZE_CONSULT" -> "尺码咨询";
            case "LOGISTICS_QUERY" -> "物流查询";
            default -> "通用咨询";
        };
    }

    private Double confidenceOf(String intentCode) {
        return switch (intentCode) {
            case "PRODUCT_DISCOUNT" -> 0.95;
            case "RETURN_EXCHANGE" -> 0.92;
            case "SIZE_CONSULT" -> 0.88;
            case "LOGISTICS_QUERY" -> 0.90;
            default -> 0.75;
        };
    }

    private String analyzeEmotionFallback(String content) {
        if (containsAny(content, "投诉", "差评", "不满", "生气", "垃圾",
                "烦", "气死", "愤怒", "恼火", "失望", "坑", "骗", "太差",
                "什么破", "无语", "崩溃", "受不了", "忍无可忍", "糟糕",
                "郁闷", "难受", "恶心", "讨厌", "离谱", "扯淡")) return "负面";
        if (containsAny(content, "感谢", "满意", "好评", "不错", "很好",
                "赞", "棒", "优秀", "喜欢", "开心", "太好了", "完美",
                "靠谱", "给力", "厉害", "贴心", "周到", "快速")) return "正面";
        return "中性";
    }

    private boolean containsAny(String content, String... keywords) {
        if (content == null) return false;
        for (String keyword : keywords) {
            if (content.contains(keyword)) return true;
        }
        return false;
    }

    private boolean matchesKeywords(String content, String commaSeparatedKeywords) {
        if (content == null || commaSeparatedKeywords == null || commaSeparatedKeywords.isBlank()) {
            return false;
        }
        return Arrays.stream(commaSeparatedKeywords.split(","))
                .map(String::trim)
                .filter(keyword -> !keyword.isEmpty())
                .anyMatch(content::contains);
    }
}
