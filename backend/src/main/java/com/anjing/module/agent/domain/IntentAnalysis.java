package com.anjing.module.agent.domain;

import lombok.Data;

/**
 * 场景、意图和情绪的统一分析结果。
 */
@Data
public class IntentAnalysis {
    private String sceneType;
    private String intentCode;
    private String intentName;
    private Double confidence;
    private String emotion;
    private AgentEngine engine;
}
