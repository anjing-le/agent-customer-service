package com.anjing.module.chat.repository;

import com.anjing.module.chat.entity.ChatRuntimeSnapshot;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ChatRuntimeSnapshotRepository extends JpaRepository<ChatRuntimeSnapshot, String> {

    List<ChatRuntimeSnapshot> findTop7ByOrderByCreatedAtDesc();
}
