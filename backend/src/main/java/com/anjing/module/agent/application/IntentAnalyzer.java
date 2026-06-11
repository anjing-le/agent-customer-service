package com.anjing.module.agent.application;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;

/**
 * 场景、意图和情绪分析端口。
 */
public interface IntentAnalyzer {
    IntentAnalysis analyze(ConversationTurn turn);
}
