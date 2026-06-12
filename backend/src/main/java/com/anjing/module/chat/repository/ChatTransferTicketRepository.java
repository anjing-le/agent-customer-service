package com.anjing.module.chat.repository;

import com.anjing.module.chat.entity.ChatTransferTicket;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.time.LocalDateTime;

@Repository
public interface ChatTransferTicketRepository extends JpaRepository<ChatTransferTicket, String> {

    List<ChatTransferTicket> findBySessionIdOrderByCreatedAtDesc(String sessionId);

    Optional<ChatTransferTicket> findFirstBySessionIdAndStatusOrderByCreatedAtDesc(String sessionId, String status);

    long countByStatus(String status);

    long countByCreatedAtBetween(LocalDateTime start, LocalDateTime end);

    long countByStatusAndResolvedAtBetween(String status, LocalDateTime start, LocalDateTime end);

    long countByStatusAndPriority(String status, String priority);

    List<ChatTransferTicket> findTop5ByOrderByCreatedAtDesc();
}
