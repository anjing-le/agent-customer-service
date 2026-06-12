package com.anjing.module.knowledge.repository;

import com.anjing.module.knowledge.entity.KnowledgeGap;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface KnowledgeGapRepository extends JpaRepository<KnowledgeGap, Long> {

    List<KnowledgeGap> findAllByOrderByUpdatedAtDesc();

    List<KnowledgeGap> findByStatusOrderByUpdatedAtDesc(String status);

    Optional<KnowledgeGap> findFirstByUserQuestionAndNoAnswerReasonAndStatusOrderByUpdatedAtDesc(
            String userQuestion,
            String noAnswerReason,
            String status
    );

    long countByStatus(String status);
}
