package com.anjing.module.agent.domain;

import lombok.Data;

import java.util.List;

/**
 * 防幻觉、安全和兜底策略的决策结果。
 */
@Data
public class GuardrailDecision {
    private boolean safe;
    private boolean fallbackRequired;
    private FallbackReason fallbackReason;
    private String userVisibleNotice;
    private List<String> policyTags;
}
