package com.anjing.module.agent.domain;

import lombok.Data;

import java.util.Map;

/**
 * 一条可被回复引用的知识证据。
 */
@Data
public class KnowledgeEvidence {
    private String evidenceId;
    private KnowledgeSource source;
    private String title;
    private String content;
    private Double score;
    private String matchReason;
    private String trustLevel;
    private Boolean quotable;
    private Map<String, Object> attributes;
}
