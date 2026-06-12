package com.anjing.module.chat;

import lombok.Data;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

/**
 * 对话模块 VO 定义
 */
public class ChatVO {

    /**
     * 对话运行概览。
     */
    @Data
    public static class RuntimeOverviewVO {
        /** 总会话数 */
        private Long totalSessions;
        /** 活跃会话数 */
        private Long activeSessions;
        /** 今日新增会话 */
        private Long todaySessions;
        /** 总消息数 */
        private Long totalMessages;
        /** 今日消息数 */
        private Long todayMessages;
        /** 今日用户消息数 */
        private Long todayUserMessages;
        /** 今日助手消息数 */
        private Long todayAssistantMessages;
        /** 今日 Agent 回复数 */
        private Long todayAgentReplies;
        /** 今日安全回复数 */
        private Long todaySafeReplies;
        /** 今日兜底回复数 */
        private Long todayFallbackReplies;
        /** 平均每会话消息数 */
        private Double averageMessagesPerSession;
        /** 最近会话 */
        private List<RecentSessionVO> recentSessions;
        /** 最近 Agent 审计 */
        private List<AgentAuditVO> recentAudits;
        /** 质检摘要 */
        private QualitySummaryVO qualitySummary;
        /** 近 7 日趋势 */
        private List<RuntimeTrendVO> dailyTrends;
        /** 最近运行快照 */
        private RuntimeSnapshotVO latestSnapshot;
    }

    /**
     * 会话级质检摘要。
     */
    @Data
    public static class SessionQualityVO {
        private Long totalAuditedReplies;
        private Long fallbackReplies;
        private Long unsafeReplies;
        private Long lowConfidenceReplies;
        private Double averageConfidence;
        private Double fallbackRate;
        private Double unsafeRate;
        private Double reliabilityScore;
        private String riskLevel;
        private String primaryFallbackReason;
    }

    /**
     * 最近会话摘要。
     */
    @Data
    public static class RecentSessionVO {
        private String sessionId;
        private String userName;
        private String channel;
        private String status;
        private Integer messageCount;
        private String lastMessage;
        private LocalDateTime updatedAt;
    }

    /**
     * Agent 单轮回复审计摘要。
     */
    @Data
    public static class AgentAuditVO {
        private String sessionId;
        private String messageId;
        private String sceneType;
        private String intentName;
        private Double confidence;
        private String replyEngine;
        private Boolean safe;
        private Boolean fallbackRequired;
        private String fallbackReason;
        private Boolean transferRecommended;
        private String transferReason;
        private String transferPriority;
        private Integer knowledgeEvidenceCount;
        private Integer ruleHitCount;
        private Integer promptRenderCount;
        private String ruleHitCodes;
        private String promptCodes;
        private LocalDateTime createdAt;
    }

    /**
     * Agent 回复质检摘要。
     */
    @Data
    public static class QualitySummaryVO {
        private Long totalAuditedReplies;
        private Double averageConfidence;
        private Double fallbackRate;
        private Double unsafeRate;
        private Double averageKnowledgeEvidenceCount;
        private Double averageRuleHitCount;
        private Double averagePromptRenderCount;
    }

    /**
     * Agent 运行趋势。
     */
    @Data
    public static class RuntimeTrendVO {
        private String date;
        private Long replies;
        private Long fallbackReplies;
        private Long unsafeReplies;
        private Double averageConfidence;
    }

    /**
     * Chat Runtime 历史快照。
     */
    @Data
    public static class RuntimeSnapshotVO {
        private String snapshotId;
        private String snapshotDate;
        private String snapshotType;
        private Long totalSessions;
        private Long activeSessions;
        private Long totalMessages;
        private Long totalAuditedReplies;
        private Double averageConfidence;
        private Double fallbackRate;
        private Double unsafeRate;
        private Double averageKnowledgeEvidenceCount;
        private Double averageRuleHitCount;
        private Double averagePromptRenderCount;
        private LocalDateTime createdAt;
    }


    /**
     * 会话信息
     */
    @Data
    public static class SessionVO {
        /** 会话ID */
        private String sessionId;
        /** 用户ID */
        private String userId;
        /** 用户名称 */
        private String userName;
        /** 会话状态 */
        private String status;
        /** 消息数量 */
        private Integer messageCount;
        /** 最后一条消息 */
        private String lastMessage;
        /** 创建时间 */
        private LocalDateTime createdAt;
    }

    /**
     * 会话详情
     */
    @Data
    public static class SessionDetailVO {
        /** 会话ID */
        private String sessionId;
        /** 用户ID */
        private String userId;
        /** 用户名称 */
        private String userName;
        /** 会话状态 */
        private String status;
        /** 上下文信息 */
        private ContextVO context;
        /** 消息列表 */
        private List<MessageVO> messages;
        /** 会话质检摘要 */
        private SessionQualityVO sessionQuality;
        /** 会话审计明细 */
        private List<AgentAuditVO> sessionAudits;
        /** 创建时间 */
        private LocalDateTime createdAt;
    }

    /**
     * 消息信息
     */
    @Data
    public static class MessageVO {
        /** 消息ID */
        private String messageId;
        /** 会话ID */
        private String sessionId;
        /** 角色：user/assistant */
        private String role;
        /** 消息内容 */
        private String content;
        /** 卡片类型 */
        private String cardType;
        /** 创建时间 */
        private LocalDateTime createdAt;
    }

    /**
     * 发送消息响应（包含AI处理结果）
     */
    @Data
    public static class SendMessageVO {
        /** 消息ID */
        private String messageId;
        /** AI回复内容 */
        private String content;
        /** 卡片类型 */
        private String cardType;
        /** 场景类型 */
        private String sceneType;
        /** 意图识别结果 */
        private IntentVO intent;
        /** 情绪分析 */
        private String emotion;
        /** 知识召回结果 */
        private KnowledgeRecallVO knowledgeRecall;
        /** 推理过程 */
        private List<ReasoningStepVO> reasoningProcess;
        /** 可靠性决策 */
        private ReliabilityVO reliability;
        /** 会话质检摘要 */
        private SessionQualityVO sessionQuality;
        /** 会话审计明细 */
        private List<AgentAuditVO> sessionAudits;
        /** 创建时间 */
        private LocalDateTime createdAt;
    }

    /**
     * 可靠性决策结果
     */
    @Data
    public static class ReliabilityVO {
        /** 回复引擎：LLM/RULE/HYBRID */
        private String replyEngine;
        /** 是否安全 */
        private Boolean safe;
        /** 是否触发兜底 */
        private Boolean fallbackRequired;
        /** 兜底原因 */
        private String fallbackReason;
        /** 用户可见提示 */
        private String userVisibleNotice;
        /** 是否建议转人工 */
        private Boolean transferRecommended;
        /** 转人工原因 */
        private String transferReason;
        /** 转人工优先级 */
        private String transferPriority;
        /** 策略标签 */
        private List<String> policyTags;
        /** 命中规则 */
        private List<RuleHitVO> ruleHits;
        /** 运行时渲染提示词 */
        private List<PromptRenderVO> promptRenders;
    }

    /**
     * 命中的运行时规则。
     */
    @Data
    public static class RuleHitVO {
        private String ruleCode;
        private String ruleName;
        private String ruleType;
        private Integer priority;
        private String reason;
        private String action;
        private String conditionSource;
    }

    /**
     * 运行时渲染后的提示词。
     */
    @Data
    public static class PromptRenderVO {
        private String promptCode;
        private String promptName;
        private String promptType;
        private String sceneType;
        private String renderedContent;
    }

    /**
     * 意图识别结果
     */
    @Data
    public static class IntentVO {
        /** 意图编码 */
        private String intentCode;
        /** 意图名称 */
        private String intentName;
        /** 置信度 */
        private Double confidence;
    }

    /**
     * 知识召回结果
     */
    @Data
    public static class KnowledgeRecallVO {
        /** 召回的商品 */
        private List<ProductRecallVO> products;
        /** 召回的FAQ */
        private List<FaqRecallVO> faqs;
        /** 召回的活动 */
        private List<ActivityRecallVO> activities;
    }

    /**
     * 商品召回
     */
    @Data
    public static class ProductRecallVO {
        /** 商品ID */
        private String productId;
        /** 商品名称 */
        private String productName;
        /** 匹配分数 */
        private Double score;
        /** 来源 */
        private String source;
    }

    /**
     * FAQ召回
     */
    @Data
    public static class FaqRecallVO {
        /** FAQ ID */
        private String faqId;
        /** 问题 */
        private String question;
        /** 答案 */
        private String answer;
        /** 匹配分数 */
        private Double score;
    }

    /**
     * 活动召回
     */
    @Data
    public static class ActivityRecallVO {
        /** 活动ID */
        private String activityId;
        /** 活动名称 */
        private String activityName;
        /** 活动描述 */
        private String description;
        /** 匹配分数 */
        private Double score;
    }

    /**
     * 推理步骤
     */
    @Data
    public static class ReasoningStepVO {
        /** 步骤序号 */
        private Integer step;
        /** 步骤标题 */
        private String title;
        /** 步骤内容 */
        private String content;
        /** 时间戳 */
        private LocalDateTime timestamp;
    }

    /**
     * 上下文信息
     */
    @Data
    public static class ContextVO {
        /** 选中的商品 */
        private List<String> selectedProducts;
        /** 选中的活动 */
        private List<String> selectedActivities;
        /** 用户画像 */
        private Map<String, Object> userProfile;
    }
}
