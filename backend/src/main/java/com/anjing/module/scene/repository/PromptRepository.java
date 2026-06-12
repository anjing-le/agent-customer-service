package com.anjing.module.scene.repository;

import com.anjing.module.scene.entity.Prompt;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface PromptRepository extends JpaRepository<Prompt, Long> {
    List<Prompt> findByStatusAndPromptType(String status, String promptType);
}
