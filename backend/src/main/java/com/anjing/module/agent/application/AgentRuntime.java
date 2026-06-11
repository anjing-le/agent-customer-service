package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.AgentReply;
import com.anjing.module.agent.domain.ConversationTurn;

/**
 * 可靠客服 Agent 主编排端口。
 */
public interface AgentRuntime {
    AgentReply handle(ConversationTurn turn);
}
