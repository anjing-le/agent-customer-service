package com.anjing.module.agent.domain;

import lombok.Data;

/**
 * 运行时提示词渲染结果。
 */
@Data
public class PromptRenderResult {
    private String promptCode;
    private String promptName;
    private String promptType;
    private String sceneType;
    private String renderedContent;
}
