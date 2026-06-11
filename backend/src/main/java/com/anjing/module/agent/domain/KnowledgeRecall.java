package com.anjing.module.agent.domain;

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

/**
 * RAG 检索阶段产出的证据集合。
 */
@Data
public class KnowledgeRecall {
    private List<KnowledgeEvidence> evidences = new ArrayList<>();
    private Double minAcceptedScore = 0.3;

    public boolean hasReliableEvidence() {
        return evidences.stream().anyMatch(e -> e.getScore() != null && e.getScore() >= minAcceptedScore);
    }
}
