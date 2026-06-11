package com.anjing.module.scene;

import com.anjing.model.constants.ApiConstants;
import com.anjing.model.response.APIResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.*;

/**
 * 场景配置控制器
 * 管理意图、提示词模板、场景规则
 */
@Slf4j
@RestController
@RequiredArgsConstructor
public class SceneController {

    private final SceneService sceneService;

    // ==================== 意图管理 ====================

    @PostMapping(ApiConstants.Scene.INTENT_LIST)
    public APIResponse<SceneVO.PageVO<SceneVO.IntentVO>> listIntents(@RequestBody SceneDTO.QueryIntentDTO dto) {
        log.info("查询意图列表: {}", dto);
        return APIResponse.success(sceneService.listIntents(dto));
    }

    @PostMapping(ApiConstants.Scene.INTENT_CREATE)
    public APIResponse<SceneVO.IntentVO> createIntent(@RequestBody SceneDTO.CreateIntentDTO dto) {
        log.info("创建意图: {}", dto);
        return APIResponse.success(sceneService.createIntent(dto));
    }

    @PostMapping(ApiConstants.Scene.INTENT_UPDATE)
    public APIResponse<SceneVO.IntentVO> updateIntent(@RequestBody SceneDTO.UpdateIntentDTO dto) {
        log.info("更新意图: {}", dto);
        return APIResponse.success(sceneService.updateIntent(dto));
    }

    @PostMapping(ApiConstants.Scene.INTENT_DELETE)
    public APIResponse<Void> deleteIntent(@RequestBody SceneDTO.IdDTO dto) {
        log.info("删除意图: {}", dto);
        sceneService.deleteIntent(dto.getId());
        return APIResponse.success();
    }

    @PostMapping(ApiConstants.Scene.INTENT_DETAIL)
    public APIResponse<SceneVO.IntentVO> getIntentDetail(@RequestBody SceneDTO.IdDTO dto) {
        log.info("获取意图详情: {}", dto);
        return APIResponse.success(sceneService.getIntentDetail(dto.getId()));
    }

    // ==================== 提示词模板 ====================

    @PostMapping(ApiConstants.Scene.PROMPT_LIST)
    public APIResponse<SceneVO.PageVO<SceneVO.PromptVO>> listPrompts(@RequestBody SceneDTO.QueryPromptDTO dto) {
        log.info("查询提示词列表: {}", dto);
        return APIResponse.success(sceneService.listPrompts(dto));
    }

    @PostMapping(ApiConstants.Scene.PROMPT_CREATE)
    public APIResponse<SceneVO.PromptVO> createPrompt(@RequestBody SceneDTO.CreatePromptDTO dto) {
        log.info("创建提示词: {}", dto);
        return APIResponse.success(sceneService.createPrompt(dto));
    }

    @PostMapping(ApiConstants.Scene.PROMPT_UPDATE)
    public APIResponse<SceneVO.PromptVO> updatePrompt(@RequestBody SceneDTO.UpdatePromptDTO dto) {
        log.info("更新提示词: {}", dto);
        return APIResponse.success(sceneService.updatePrompt(dto));
    }

    @PostMapping(ApiConstants.Scene.PROMPT_DELETE)
    public APIResponse<Void> deletePrompt(@RequestBody SceneDTO.IdDTO dto) {
        log.info("删除提示词: {}", dto);
        sceneService.deletePrompt(dto.getId());
        return APIResponse.success();
    }

    @PostMapping(ApiConstants.Scene.PROMPT_DETAIL)
    public APIResponse<SceneVO.PromptVO> getPromptDetail(@RequestBody SceneDTO.IdDTO dto) {
        log.info("获取提示词详情: {}", dto);
        return APIResponse.success(sceneService.getPromptDetail(dto.getId()));
    }

    @PostMapping(ApiConstants.Scene.PROMPT_TEST)
    public APIResponse<SceneVO.PromptTestResultVO> testPrompt(@RequestBody SceneDTO.TestPromptDTO dto) {
        log.info("测试提示词: {}", dto);
        return APIResponse.success(sceneService.testPrompt(dto));
    }

    // ==================== 场景规则 ====================

    @PostMapping(ApiConstants.Scene.RULE_LIST)
    public APIResponse<SceneVO.PageVO<SceneVO.RuleVO>> listRules(@RequestBody SceneDTO.QueryRuleDTO dto) {
        log.info("查询规则列表: {}", dto);
        return APIResponse.success(sceneService.listRules(dto));
    }

    @PostMapping(ApiConstants.Scene.RULE_CREATE)
    public APIResponse<SceneVO.RuleVO> createRule(@RequestBody SceneDTO.CreateRuleDTO dto) {
        log.info("创建规则: {}", dto);
        return APIResponse.success(sceneService.createRule(dto));
    }

    @PostMapping(ApiConstants.Scene.RULE_UPDATE)
    public APIResponse<SceneVO.RuleVO> updateRule(@RequestBody SceneDTO.UpdateRuleDTO dto) {
        log.info("更新规则: {}", dto);
        return APIResponse.success(sceneService.updateRule(dto));
    }

    @PostMapping(ApiConstants.Scene.RULE_DELETE)
    public APIResponse<Void> deleteRule(@RequestBody SceneDTO.IdDTO dto) {
        log.info("删除规则: {}", dto);
        sceneService.deleteRule(dto.getId());
        return APIResponse.success();
    }

    @PostMapping(ApiConstants.Scene.RULE_DETAIL)
    public APIResponse<SceneVO.RuleVO> getRuleDetail(@RequestBody SceneDTO.IdDTO dto) {
        log.info("获取规则详情: {}", dto);
        return APIResponse.success(sceneService.getRuleDetail(dto.getId()));
    }

    @PostMapping(ApiConstants.Scene.RULE_ENABLE)
    public APIResponse<Void> enableRule(@RequestBody SceneDTO.IdDTO dto) {
        log.info("启用规则: {}", dto);
        sceneService.enableRule(dto.getId());
        return APIResponse.success();
    }

    @PostMapping(ApiConstants.Scene.RULE_DISABLE)
    public APIResponse<Void> disableRule(@RequestBody SceneDTO.IdDTO dto) {
        log.info("禁用规则: {}", dto);
        sceneService.disableRule(dto.getId());
        return APIResponse.success();
    }
}

