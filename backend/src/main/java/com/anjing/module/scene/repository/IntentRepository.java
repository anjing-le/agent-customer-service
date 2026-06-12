package com.anjing.module.scene.repository;

import com.anjing.module.scene.entity.Intent;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface IntentRepository extends JpaRepository<Intent, Long> {
    List<Intent> findByStatusOrderByPriorityAsc(String status);
}
