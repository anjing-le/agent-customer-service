package com.anjing.module.chat.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * Agent 单轮回复审计。
 */
@Data
@Entity
@Table(name = "cs_chat_agent_audit")
public class ChatAgentAudit {

    @Id
    @Column(length = 64)
    private String auditId;

    @Column(length = 64, nullable = false)
    private String sessionId;

    @Column(length = 64, nullable = false)
    private String messageId;

    @Column(length = 64)
    private String sceneType;

    @Column(length = 64)
    private String intentCode;

    @Column(length = 64)
    private String intentName;

    private Double confidence;

    @Column(length = 32)
    private String replyEngine;

    private Boolean safe;

    private Boolean fallbackRequired;

    @Column(length = 64)
    private String fallbackReason;

    private Integer knowledgeEvidenceCount;

    private Integer ruleHitCount;

    private Integer promptRenderCount;

    @Column(length = 512)
    private String ruleHitCodes;

    @Column(length = 512)
    private String promptCodes;

    private LocalDateTime createdAt;

    @PrePersist
    public void prePersist() {
        if (createdAt == null) createdAt = LocalDateTime.now();
    }
}
