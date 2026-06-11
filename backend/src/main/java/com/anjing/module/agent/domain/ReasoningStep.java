package com.anjing.module.agent.domain;

import lombok.Data;

import java.time.LocalDateTime;

/**
 * 可审计的 Agent 推理步骤。
 */
@Data
public class ReasoningStep {
    private Integer step;
    private String title;
    private String content;
    private LocalDateTime timestamp;
}
