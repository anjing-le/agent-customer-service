package com.anjing.module.agent.domain;

import lombok.Data;

import java.util.List;
import java.util.Map;

/**
 * 一次用户输入触发的 Agent 处理回合。
 */
@Data
public class ConversationTurn {
    private String turnId;
    private String sessionId;
    private String userId;
    private String channel;
    private String userMessage;
    private List<ConversationMessage> recentHistory;
    private Map<String, Object> context;
}
