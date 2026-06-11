/**
 * 知识管理 API
 * 路径与后端 KnowledgeController (ApiConstants.Knowledge) 完全对齐
 */
import request from '@/utils/http'
import { ApiPaths } from '@/api/apiPaths'

// ==================== 知识总览 ====================

/**
 * 获取知识总览
 * POST /api/knowledge/overview
 */
export function fetchGetOverview() {
  return request.post<any>({
    url: ApiPaths.knowledge.overview,
    data: {}
  })
}

// ==================== 商品知识 ====================

export function fetchListProducts(data?: { keyword?: string; category?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.productList,
    data: data || {}
  })
}

export function fetchCreateProduct(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.productCreate,
    data
  })
}

export function fetchUpdateProduct(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.productUpdate,
    data
  })
}

export function fetchDeleteProduct(id: number) {
  return request.post<void>({
    url: ApiPaths.knowledge.productDelete,
    data: { id }
  })
}

export function fetchGetProductDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.knowledge.productDetail,
    data: { id }
  })
}

// ==================== 活动知识 ====================

export function fetchListActivities(data?: { keyword?: string; activityType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.activityList,
    data: data || {}
  })
}

export function fetchCreateActivity(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.activityCreate,
    data
  })
}

export function fetchUpdateActivity(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.activityUpdate,
    data
  })
}

export function fetchDeleteActivity(id: number) {
  return request.post<void>({
    url: ApiPaths.knowledge.activityDelete,
    data: { id }
  })
}

export function fetchGetActivityDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.knowledge.activityDetail,
    data: { id }
  })
}

// ==================== FAQ问答 ====================

export function fetchListFaqs(data?: { keyword?: string; category?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.faqList,
    data: data || {}
  })
}

export function fetchCreateFaq(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.faqCreate,
    data
  })
}

export function fetchUpdateFaq(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.faqUpdate,
    data
  })
}

export function fetchDeleteFaq(id: number) {
  return request.post<void>({
    url: ApiPaths.knowledge.faqDelete,
    data: { id }
  })
}

export function fetchGetFaqDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.knowledge.faqDetail,
    data: { id }
  })
}

// ==================== 行业知识 ====================

export function fetchListIndustries(data?: { keyword?: string; industryType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.industryList,
    data: data || {}
  })
}

export function fetchCreateIndustry(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.industryCreate,
    data
  })
}

export function fetchUpdateIndustry(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.industryUpdate,
    data
  })
}

export function fetchDeleteIndustry(id: number) {
  return request.post<void>({
    url: ApiPaths.knowledge.industryDelete,
    data: { id }
  })
}

// ==================== 场景解决方案 ====================

export function fetchListSolutions(data?: { keyword?: string; sceneType?: string; status?: string; page?: number; size?: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.solutionList,
    data: data || {}
  })
}

export function fetchCreateSolution(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.solutionCreate,
    data
  })
}

export function fetchUpdateSolution(data: any) {
  return request.post<any>({
    url: ApiPaths.knowledge.solutionUpdate,
    data
  })
}

export function fetchDeleteSolution(id: number) {
  return request.post<void>({
    url: ApiPaths.knowledge.solutionDelete,
    data: { id }
  })
}

export function fetchGetSolutionDetail(id: number) {
  return request.post<any>({
    url: ApiPaths.knowledge.solutionDetail,
    data: { id }
  })
}

// ==================== 知识向量化 ====================

export function fetchVectorize(data: { knowledgeType: string; knowledgeId: number }) {
  return request.post<any>({
    url: ApiPaths.knowledge.vectorize,
    data
  })
}

export function fetchGetVectorizeStatus(taskId: string) {
  return request.post<any>({
    url: ApiPaths.knowledge.vectorizeStatus,
    data: { taskId }
  })
}

// ==================== 导入导出 ====================

export function fetchImportProducts(data: { knowledgeType: string; fileUrl: string }) {
  return request.post<any>({
    url: ApiPaths.knowledge.productImport,
    data
  })
}

export function fetchExportProducts(data: { knowledgeType: string; ids?: number[] }) {
  return request.post<any>({
    url: ApiPaths.knowledge.productExport,
    data
  })
}

export function fetchImportFaqs(data: { knowledgeType: string; fileUrl: string }) {
  return request.post<any>({
    url: ApiPaths.knowledge.faqImport,
    data
  })
}
