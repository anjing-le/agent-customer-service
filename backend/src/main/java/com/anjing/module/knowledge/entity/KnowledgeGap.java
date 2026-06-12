package com.anjing.module.knowledge.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 对话运行时暴露出来的知识缺口。
 */
@Data
@Entity
@Table(name = "cs_knowledge_gap")
public class KnowledgeGap {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(length = 64)
    private String sessionId;

    @Column(length = 64)
    private String messageId;

    @Column(length = 512, nullable = false)
    private String userQuestion;

    @Column(length = 64)
    private String intentCode;

    @Column(length = 128)
    private String intentName;

    @Column(length = 64)
    private String sceneType;

    @Column(length = 64)
    private String noAnswerReason;

    @Column(length = 512)
    private String noAnswerDetail;

    @Column(length = 32)
    private String status;

    @Column(length = 32)
    private String priority;

    private Integer occurrenceCount;

    @Column(length = 64)
    private String resolvedKnowledgeType;

    private Long resolvedKnowledgeId;

    @Column(length = 512)
    private String resolutionNote;

    private LocalDateTime firstSeenAt;
    private LocalDateTime lastSeenAt;
    private LocalDateTime resolvedAt;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    @PrePersist
    public void prePersist() {
        LocalDateTime now = LocalDateTime.now();
        if (createdAt == null) createdAt = now;
        if (updatedAt == null) updatedAt = now;
        if (firstSeenAt == null) firstSeenAt = now;
        if (lastSeenAt == null) lastSeenAt = now;
        if (status == null) status = "OPEN";
        if (priority == null) priority = "MEDIUM";
        if (occurrenceCount == null) occurrenceCount = 1;
    }

    @PreUpdate
    public void preUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
