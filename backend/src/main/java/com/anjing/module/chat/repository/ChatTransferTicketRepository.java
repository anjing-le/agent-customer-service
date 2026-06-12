package com.anjing.module.chat.repository;

import com.anjing.module.chat.entity.ChatTransferTicket;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ChatTransferTicketRepository extends JpaRepository<ChatTransferTicket, String> {

    List<ChatTransferTicket> findBySessionIdOrderByCreatedAtDesc(String sessionId);

    Optional<ChatTransferTicket> findFirstBySessionIdAndStatusOrderByCreatedAtDesc(String sessionId, String status);
}
