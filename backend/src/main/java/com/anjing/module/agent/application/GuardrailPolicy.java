package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;

/**
 * 防幻觉、安全和兜底策略端口。
 */
public interface GuardrailPolicy {
    GuardrailDecision decide(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall);
}
