package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.AgentReply;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;

/**
 * LLM 回复生成与规则兜底端口。
 */
public interface ReplyGenerator {
    AgentReply generate(
            ConversationTurn turn,
            IntentAnalysis analysis,
            KnowledgeRecall recall,
            GuardrailDecision guardrailDecision
    );
}
