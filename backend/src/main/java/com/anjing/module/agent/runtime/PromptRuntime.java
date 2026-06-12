package com.anjing.module.agent.runtime;

import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.PromptRenderResult;
import com.anjing.module.scene.entity.Prompt;
import com.anjing.module.scene.repository.PromptRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Prompt 运行时：筛选启用模板并渲染运行时变量。
 */
@Service
@RequiredArgsConstructor
public class PromptRuntime {

    private final PromptRepository promptRepository;

    public List<PromptRenderResult> renderSystemPrompts(
            ConversationTurn turn,
            IntentAnalysis analysis,
            KnowledgeRecall recall
    ) {
        List<PromptRenderResult> results = new ArrayList<>();
        Map<String, String> variables = buildVariables(turn, analysis, recall);

        for (Prompt prompt : promptRepository.findByStatusAndPromptType("启用", "SYSTEM")) {
            if (!matchesScene(prompt, analysis)) continue;
            if (prompt.getPromptContent() == null || prompt.getPromptContent().isBlank()) continue;

            PromptRenderResult result = new PromptRenderResult();
            result.setPromptCode(prompt.getPromptCode());
            result.setPromptName(prompt.getPromptName());
            result.setPromptType(prompt.getPromptType());
            result.setSceneType(prompt.getSceneType());
            result.setRenderedContent(render(prompt.getPromptContent(), variables));
            results.add(result);

            prompt.setUsageCount(prompt.getUsageCount() == null ? 1 : prompt.getUsageCount() + 1);
            promptRepository.save(prompt);
        }
        return results;
    }

    public String joinRenderedContent(List<PromptRenderResult> results) {
        if (results == null || results.isEmpty()) return null;
        StringBuilder builder = new StringBuilder();
        for (PromptRenderResult result : results) {
            if (builder.length() > 0) builder.append("\n");
            builder.append("- ")
                    .append(result.getPromptName())
                    .append(": ")
                    .append(result.getRenderedContent());
        }
        return builder.toString();
    }

    private Map<String, String> buildVariables(ConversationTurn turn, IntentAnalysis analysis, KnowledgeRecall recall) {
        Map<String, String> variables = new HashMap<>();
        variables.put("sceneType", nullToEmpty(analysis.getSceneType()));
        variables.put("intentCode", nullToEmpty(analysis.getIntentCode()));
        variables.put("intentName", nullToEmpty(analysis.getIntentName()));
        variables.put("confidence", analysis.getConfidence() != null ? String.valueOf(analysis.getConfidence()) : "");
        variables.put("emotion", nullToEmpty(analysis.getEmotion()));
        variables.put("userMessage", nullToEmpty(turn.getUserMessage()));
        variables.put("knowledgeCount", String.valueOf(recall.getEvidences().size()));
        variables.put("hasReliableKnowledge", String.valueOf(recall.hasReliableEvidence()));
        return variables;
    }

    private String render(String template, Map<String, String> variables) {
        String rendered = template;
        for (Map.Entry<String, String> entry : variables.entrySet()) {
            rendered = rendered.replace("{{" + entry.getKey() + "}}", entry.getValue());
        }
        return rendered;
    }

    private boolean matchesScene(Prompt prompt, IntentAnalysis analysis) {
        return prompt.getSceneType() == null
                || prompt.getSceneType().isBlank()
                || prompt.getSceneType().equals(analysis.getSceneType());
    }

    private String nullToEmpty(String value) {
        return value == null ? "" : value;
    }
}
