package com.anjing.module.agent.domain;

/**
 * 触发规则兜底或转人工的原因。
 */
public enum FallbackReason {
    LLM_UNAVAILABLE,
    LOW_CONFIDENCE,
    NO_RELIABLE_KNOWLEDGE,
    SAFETY_BLOCKED,
    UNSUPPORTED_INTENT
}
