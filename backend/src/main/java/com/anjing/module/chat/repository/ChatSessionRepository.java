package com.anjing.module.chat.repository;

import com.anjing.module.chat.entity.ChatSession;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface ChatSessionRepository extends JpaRepository<ChatSession, String> {

    List<ChatSession> findByUserIdOrderByCreatedAtDesc(String userId);

    List<ChatSession> findByStatusOrderByCreatedAtDesc(String status);

    List<ChatSession> findByUserIdAndStatusOrderByCreatedAtDesc(String userId, String status);

    List<ChatSession> findAllByOrderByCreatedAtDesc();

    long countByStatus(String status);

    long countByCreatedAtBetween(LocalDateTime start, LocalDateTime end);
}
