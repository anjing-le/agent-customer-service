package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.AgentRuntime;
import com.anjing.module.agent.application.GuardrailPolicy;
import com.anjing.module.agent.application.IntentAnalyzer;
import com.anjing.module.agent.application.KnowledgeRetriever;
import com.anjing.module.agent.application.ReplyGenerator;
import com.anjing.module.agent.domain.AgentReply;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.GuardrailDecision;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.ReasoningStep;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

/**
 * 可靠客服 Agent 的主编排实现。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class DefaultAgentRuntime implements AgentRuntime {

    private final IntentAnalyzer intentAnalyzer;
    private final KnowledgeRetriever knowledgeRetriever;
    private final GuardrailPolicy guardrailPolicy;
    private final ReplyGenerator replyGenerator;

    @Override
    public AgentReply handle(ConversationTurn turn) {
        log.info("========== Agent Runtime 开始 ==========");
        IntentAnalysis analysis = intentAnalyzer.analyze(turn);
        KnowledgeRecall recall = knowledgeRetriever.retrieve(turn, analysis);
        GuardrailDecision guardrailDecision = guardrailPolicy.decide(turn, analysis, recall);
        AgentReply reply = replyGenerator.generate(turn, analysis, recall, guardrailDecision);
        reply.setReasoningSteps(buildReasoningSteps(analysis, recall, guardrailDecision, reply));
        log.info("========== Agent Runtime 结束 | scene={} | intent={} | replyEngine={} | fallback={} ==========",
                analysis.getSceneType(), analysis.getIntentName(), reply.getEngine(), guardrailDecision.isFallbackRequired());
        return reply;
    }

    private List<ReasoningStep> buildReasoningSteps(
            IntentAnalysis analysis,
            KnowledgeRecall recall,
            GuardrailDecision guardrailDecision,
            AgentReply reply
    ) {
        List<ReasoningStep> steps = new ArrayList<>();
        addStep(steps, 1, "输入解析", "分析引擎：" + analysis.getEngine());
        addStep(steps, 2, "场景识别", "识别场景类型：" + analysis.getSceneType());
        addStep(steps, 3, "意图分析", "意图：" + analysis.getIntentName() + "，置信度：" + analysis.getConfidence());
        addStep(steps, 4, "情绪判断", "用户情绪：" + analysis.getEmotion());
        addStep(steps, 5, "知识检索", "检索到证据：" + recall.getEvidences().size() + "条，可靠证据：" + recall.hasReliableEvidence());
        addStep(steps, 6, "可靠性护栏", guardrailDecision.isFallbackRequired()
                ? "触发兜底：" + guardrailDecision.getFallbackReason()
                : "未触发兜底，允许生成回复");
        addStep(steps, 7, "生成回复", "回复引擎：" + reply.getEngine());
        return steps;
    }

    private void addStep(List<ReasoningStep> steps, int step, String title, String content) {
        ReasoningStep reasoningStep = new ReasoningStep();
        reasoningStep.setStep(step);
        reasoningStep.setTitle(title);
        reasoningStep.setContent(content);
        reasoningStep.setTimestamp(LocalDateTime.now());
        steps.add(reasoningStep);
    }
}
