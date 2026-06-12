package com.anjing.module.chat.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 转人工队列工单。
 */
@Data
@Entity
@Table(name = "cs_chat_transfer_ticket")
public class ChatTransferTicket {

    @Id
    @Column(length = 64)
    private String ticketId;

    @Column(length = 64, nullable = false)
    private String sessionId;

    @Column(length = 64)
    private String messageId;

    @Column(length = 32)
    private String status;

    @Column(length = 32)
    private String priority;

    @Column(length = 512)
    private String reason;

    @Column(length = 64)
    private String assignedAgentId;

    @Column(length = 64)
    private String assignedAgentName;

    @Column(length = 512)
    private String resolutionNote;

    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    private LocalDateTime resolvedAt;

    @PrePersist
    public void prePersist() {
        if (createdAt == null) createdAt = LocalDateTime.now();
        if (updatedAt == null) updatedAt = LocalDateTime.now();
        if (status == null) status = "PENDING";
    }

    @PreUpdate
    public void preUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
