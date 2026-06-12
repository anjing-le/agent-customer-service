package com.anjing.module.chat.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;
import lombok.Data;

import java.time.LocalDate;
import java.time.LocalDateTime;

/**
 * Chat Runtime 运行快照。
 */
@Data
@Entity
@Table(name = "cs_chat_runtime_snapshot")
public class ChatRuntimeSnapshot {

    @Id
    @Column(length = 64)
    private String snapshotId;

    private LocalDate snapshotDate;

    @Column(length = 32)
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

    @PrePersist
    public void prePersist() {
        if (createdAt == null) createdAt = LocalDateTime.now();
    }
}
