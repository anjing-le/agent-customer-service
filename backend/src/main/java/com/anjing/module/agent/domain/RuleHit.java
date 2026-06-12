package com.anjing.module.agent.domain;

import lombok.Data;

/**
 * 运行时命中的场景规则。
 */
@Data
public class RuleHit {
    private String ruleCode;
    private String ruleName;
    private String ruleType;
    private Integer priority;
    private String reason;
    private String action;
}
