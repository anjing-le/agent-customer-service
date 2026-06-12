package com.anjing.module.chat.repository;

import com.anjing.module.chat.entity.ChatAgentAudit;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface ChatAgentAuditRepository extends JpaRepository<ChatAgentAudit, String> {

    long countByCreatedAtBetween(LocalDateTime start, LocalDateTime end);

    long countBySafeAndCreatedAtBetween(Boolean safe, LocalDateTime start, LocalDateTime end);

    long countByFallbackRequiredAndCreatedAtBetween(Boolean fallbackRequired, LocalDateTime start, LocalDateTime end);

    List<ChatAgentAudit> findByCreatedAtBetweenOrderByCreatedAtAsc(LocalDateTime start, LocalDateTime end);

    List<ChatAgentAudit> findBySessionIdOrderByCreatedAtAsc(String sessionId);

    List<ChatAgentAudit> findTop5ByOrderByCreatedAtDesc();
}
