package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;

/**
 * 知识检索端口。
 */
public interface KnowledgeRetriever {
    KnowledgeRecall retrieve(ConversationTurn turn, IntentAnalysis analysis);
}
