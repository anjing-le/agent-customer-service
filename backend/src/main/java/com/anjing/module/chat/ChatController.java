package com.anjing.module.chat;

import com.anjing.model.constants.ApiConstants;
import com.anjing.model.response.APIResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 对话中心控制器
 * 核心对话功能，包含会话管理和消息处理
 */
@Slf4j
@RestController
@RequiredArgsConstructor
public class ChatController {

    private final ChatService chatService;

    /**
     * 获取对话运行概览
     */
    @PostMapping(ApiConstants.Chat.RUNTIME_OVERVIEW)
    public APIResponse<ChatVO.RuntimeOverviewVO> getRuntimeOverview() {
        return APIResponse.success(chatService.getRuntimeOverview());
    }

    // ==================== 会话管理 ====================

    /**
     * 创建会话
     */
    @PostMapping(ApiConstants.Chat.CREATE_SESSION)
    public APIResponse<ChatVO.SessionVO> createSession(@RequestBody ChatDTO.CreateSessionDTO dto) {
        log.info("创建会话: {}", dto);
        return APIResponse.success(chatService.createSession(dto));
    }

    /**
     * 获取会话列表
     */
    @PostMapping(ApiConstants.Chat.LIST_SESSIONS)
    public APIResponse<List<ChatVO.SessionVO>> listSessions(@RequestBody ChatDTO.QuerySessionDTO dto) {
        log.info("查询会话列表: {}", dto);
        return APIResponse.success(chatService.listSessions(dto));
    }

    /**
     * 获取会话详情
     */
    @PostMapping(ApiConstants.Chat.GET_SESSION)
    public APIResponse<ChatVO.SessionDetailVO> getSession(@RequestBody ChatDTO.SessionIdDTO dto) {
        log.info("获取会话详情: {}", dto);
        return APIResponse.success(chatService.getSession(dto.getSessionId()));
    }

    /**
     * 删除会话
     */
    @PostMapping(ApiConstants.Chat.DELETE_SESSION)
    public APIResponse<Void> deleteSession(@RequestBody ChatDTO.SessionIdDTO dto) {
        log.info("删除会话: {}", dto);
        chatService.deleteSession(dto.getSessionId());
        return APIResponse.success();
    }

    // ==================== 消息处理（核心链路） ====================

    /**
     * 发送消息（核心接口）
     * 用户发送消息 -> 意图识别 -> 知识检索 -> LLM生成回复 -> 返回响应
     */
    @PostMapping(ApiConstants.Chat.SEND_MESSAGE)
    public APIResponse<ChatVO.SendMessageVO> sendMessage(@RequestBody ChatDTO.SendMessageDTO dto) {
        log.info("发送消息: sessionId={}, content={}", dto.getSessionId(), dto.getContent());
        return APIResponse.success(chatService.sendMessage(dto));
    }

    /**
     * 获取会话消息历史
     */
    @PostMapping(ApiConstants.Chat.GET_MESSAGES)
    public APIResponse<List<ChatVO.MessageVO>> getMessages(@RequestBody ChatDTO.QueryMessagesDTO dto) {
        log.info("获取消息历史: {}", dto);
        return APIResponse.success(chatService.getMessages(dto));
    }

    // ==================== 上下文管理 ====================

    /**
     * 更新会话上下文
     */
    @PostMapping(ApiConstants.Chat.UPDATE_CONTEXT)
    public APIResponse<Void> updateContext(@RequestBody ChatDTO.UpdateContextDTO dto) {
        log.info("更新上下文: {}", dto);
        chatService.updateContext(dto);
        return APIResponse.success();
    }

    /**
     * 获取会话上下文
     */
    @PostMapping(ApiConstants.Chat.GET_CONTEXT)
    public APIResponse<ChatVO.ContextVO> getContext(@RequestBody ChatDTO.SessionIdDTO dto) {
        log.info("获取上下文: {}", dto);
        return APIResponse.success(chatService.getContext(dto.getSessionId()));
    }
}
