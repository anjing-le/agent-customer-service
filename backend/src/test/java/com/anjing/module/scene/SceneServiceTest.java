package com.anjing.module.scene;

import com.anjing.model.exception.BizException;
import com.anjing.module.chat.LlmService;
import com.anjing.module.agent.runtime.RuleEngine;
import com.anjing.module.scene.repository.IntentRepository;
import com.anjing.module.scene.repository.PromptRepository;
import com.anjing.module.scene.repository.RuleRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.mock;

class SceneServiceTest {

    private final SceneService sceneService = new SceneService(
            mock(IntentRepository.class),
            mock(PromptRepository.class),
            mock(RuleRepository.class),
            mock(LlmService.class),
            new ObjectMapper(),
            mock(RuleEngine.class)
    );

    @Test
    void rejectsRequiredPromptVariableMissingFromTemplate() {
        SceneDTO.CreatePromptDTO dto = new SceneDTO.CreatePromptDTO();
        dto.setPromptName("售后安抚");
        dto.setPromptCode("AFTER_SALE_COMFORT");
        dto.setPromptType("SYSTEM");
        dto.setContent("请根据用户诉求生成客服回复");

        SceneDTO.PromptVariableDTO variable = new SceneDTO.PromptVariableDTO();
        variable.setName("intentName");
        variable.setDescription("意图名称");
        variable.setRequired(true);
        dto.setVariables(List.of(variable));

        assertThatThrownBy(() -> sceneService.createPrompt(dto))
                .isInstanceOf(BizException.class)
                .hasMessageContaining("必填 Prompt 变量未出现在模板中");
    }
}
