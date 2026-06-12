/**
 * 场景配置 API
 * 路径与后端 SceneController (ApiConstants.Scene) 完全对齐
 */
import request from '@/utils/http'
import { ApiPaths } from '@/api/apiPaths'

export function fetchSceneRuntimeOverview() {
  return request.post<any>({
    url: ApiPaths.scene.runtimeOverview,
    data: {}
  })
}

// ==================== 意图管理 ====================

export function fetchListIntents(data?: { keyword?: string; sceneType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.scene.intentList,
    data: data || {}
  })
}

export function fetchCreateIntent(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.intentCreate,
    data
  })
}

export function fetchUpdateIntent(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.intentUpdate,
    data
  })
}

export function fetchDeleteIntent(id: number) {
  return request.post<void>({
    url: ApiPaths.scene.intentDelete,
    data: { id }
  })
}

export function fetchGetIntentDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.scene.intentDetail,
    data: { id }
  })
}

// ==================== 提示词模板 ====================

export function fetchListPrompts(data?: { keyword?: string; sceneType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.scene.promptList,
    data: data || {}
  })
}

export function fetchCreatePrompt(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.promptCreate,
    data
  })
}

export function fetchUpdatePrompt(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.promptUpdate,
    data
  })
}

export function fetchDeletePrompt(id: number) {
  return request.post<void>({
    url: ApiPaths.scene.promptDelete,
    data: { id }
  })
}

export function fetchGetPromptDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.scene.promptDetail,
    data: { id }
  })
}

export function fetchTestPrompt(data: { promptId: number; input?: string; variables?: Record<string, string> }) {
  return request.post<any>({
    url: ApiPaths.scene.promptTest,
    data
  })
}

// ==================== 场景规则 ====================

export function fetchListRules(data?: { keyword?: string; sceneType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.scene.ruleList,
    data: data || {}
  })
}

export function fetchCreateRule(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.ruleCreate,
    data
  })
}

export function fetchUpdateRule(data: any) {
  return request.post<any>({
    url: ApiPaths.scene.ruleUpdate,
    data
  })
}

export function fetchDeleteRule(id: number) {
  return request.post<void>({
    url: ApiPaths.scene.ruleDelete,
    data: { id }
  })
}

export function fetchGetRuleDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.scene.ruleDetail,
    data: { id }
  })
}

export function fetchEnableRule(id: number) {
  return request.post<void>({
    url: ApiPaths.scene.ruleEnable,
    data: { id }
  })
}

export function fetchDisableRule(id: number) {
  return request.post<void>({
    url: ApiPaths.scene.ruleDisable,
    data: { id }
  })
}

export function fetchTestRule(data: {
  ruleId: number
  userMessage?: string
  sceneType?: string
  intentCode?: string
  intentName?: string
  confidence?: number
  emotion?: string
  knowledgeCount?: number
  hasReliableKnowledge?: boolean
  context?: Record<string, any>
}) {
  return request.post<any>({
    url: ApiPaths.scene.ruleTest,
    data
  })
}
