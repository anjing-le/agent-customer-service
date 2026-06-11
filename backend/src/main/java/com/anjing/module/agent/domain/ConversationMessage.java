package com.anjing.module.agent.domain;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.Map;

/**
 * 会话内的一条消息。
 */
@Data
public class ConversationMessage {
    private String messageId;
    private ConversationRole role;
    private String content;
    private LocalDateTime createdAt;
    private Map<String, Object> metadata;
}
