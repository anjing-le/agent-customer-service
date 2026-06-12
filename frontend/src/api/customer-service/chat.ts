/**
 * 对话管理 API
 * 路径与后端 ChatController (ApiConstants.Chat) 完全对齐
 */
import request from '@/utils/http'
import { ApiPaths } from '@/api/apiPaths'

export function fetchChatRuntimeOverview() {
  return request.post<any>({
    url: ApiPaths.chat.runtimeOverview,
    data: {}
  })
}

export function fetchCaptureChatRuntimeSnapshot() {
  return request.post<any>({
    url: ApiPaths.chat.runtimeSnapshotCapture,
    data: {}
  })
}

// ==================== 会话管理 ====================

/**
 * 创建会话
 * POST /api/chat/session/create
 */
export function fetchCreateSession(data: { userId: string; userName?: string; channel?: string }) {
  return request.post<any>({
    url: ApiPaths.chat.createSession,
    data
  })
}

/**
 * 获取会话列表
 * POST /api/chat/session/list
 */
export function fetchGetSessionList(data?: { userId?: string; status?: string }) {
  return request.post<any[]>({
    url: ApiPaths.chat.listSessions,
    data: data || {}
  })
}

/**
 * 获取会话详情
 * POST /api/chat/session/get
 */
export function fetchGetSessionDetail(sessionId: string) {
  return request.post<any>({
    url: ApiPaths.chat.getSession,
    data: { sessionId }
  })
}

/**
 * 删除会话
 * POST /api/chat/session/delete
 */
export function fetchDeleteSession(sessionId: string) {
  return request.post<void>({
    url: ApiPaths.chat.deleteSession,
    data: { sessionId }
  })
}

// ==================== 消息处理（核心链路） ====================

/**
 * 发送消息
 * POST /api/chat/message/send
 */
export function fetchSendMessage(data: { sessionId: string; content: string; contentType?: string; extra?: Record<string, any> }) {
  return request.post<any>({
    url: ApiPaths.chat.sendMessage,
    data
  })
}

/**
 * 获取会话消息历史
 * POST /api/chat/message/list
 */
export function fetchGetMessageList(data: { sessionId: string; page?: number; size?: number }) {
  return request.post<any[]>({
    url: ApiPaths.chat.listMessages,
    data
  })
}

// ==================== 上下文管理 ====================

/**
 * 更新会话上下文
 * POST /api/chat/context/update
 */
export function fetchUpdateContext(data: {
  sessionId: string
  selectedProducts?: string[]
  selectedActivities?: string[]
  userProfile?: Record<string, any>
}) {
  return request.post<void>({
    url: ApiPaths.chat.updateContext,
    data
  })
}

/**
 * 获取会话上下文
 * POST /api/chat/context/get
 */
export function fetchGetContext(sessionId: string) {
  return request.post<any>({
    url: ApiPaths.chat.getContext,
    data: { sessionId }
  })
}

// ==================== 转人工流程 ====================

export function fetchListTransferTickets(data?: { sessionId?: string; status?: string }) {
  return request.post<any[]>({
    url: ApiPaths.chat.listTransferTickets,
    data: data || {}
  })
}

export function fetchResolveTransferTicket(data: {
  ticketId: string
  agentId?: string
  agentName?: string
  resolutionNote?: string
}) {
  return request.post<any>({
    url: ApiPaths.chat.resolveTransferTicket,
    data
  })
}
