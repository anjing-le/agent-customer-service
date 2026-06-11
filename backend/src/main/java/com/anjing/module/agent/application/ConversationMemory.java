package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.ConversationMessage;
import com.anjing.module.agent.domain.ConversationTurn;

import java.util.List;

/**
 * 会话历史与上下文端口。
 */
public interface ConversationMemory {
    List<ConversationMessage> loadRecentHistory(String sessionId, int limit);

    void appendUserTurn(ConversationTurn turn);
}
