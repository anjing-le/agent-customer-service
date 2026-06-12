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
    private Boolean answerable = true;
    private String noAnswerReason;
    private String noAnswerDetail;
    private Boolean hallucinationBlocked = false;

    public boolean hasReliableEvidence() {
        return evidences.stream().anyMatch(e ->
                Boolean.TRUE.equals(e.getQuotable())
                        && e.getScore() != null
                        && e.getScore() >= minAcceptedScore
        );
    }
}
