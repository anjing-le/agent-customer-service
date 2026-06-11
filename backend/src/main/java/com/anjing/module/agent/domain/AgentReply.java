package com.anjing.module.agent.domain;

import lombok.Data;

import java.util.List;

/**
 * 一轮 Agent 回复的完整结果。
 */
@Data
public class AgentReply {
    private String messageId;
    private String content;
    private AgentEngine engine;
    private String cardType;
    private IntentAnalysis intentAnalysis;
    private KnowledgeRecall knowledgeRecall;
    private GuardrailDecision guardrailDecision;
    private List<ReasoningStep> reasoningSteps;
}
