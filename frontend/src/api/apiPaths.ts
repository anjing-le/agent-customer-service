/**
 * Runtime API paths.
 *
 * V1 keeps the existing `/api/chat`, `/api/knowledge` and `/api/scene`
 * boundaries. Future migration to `/api/customer-service/**` should start
 * from `contracts/service-boundaries.json`, then update this file.
 */
export const ApiPaths = {
  chat: {
    runtimeOverview: '/api/chat/runtime/overview',
    runtimeSnapshotCapture: '/api/chat/runtime/snapshot/capture',
    createSession: '/api/chat/session/create',
    listSessions: '/api/chat/session/list',
    getSession: '/api/chat/session/get',
    deleteSession: '/api/chat/session/delete',
    sendMessage: '/api/chat/message/send',
    listMessages: '/api/chat/message/list',
    updateContext: '/api/chat/context/update',
    getContext: '/api/chat/context/get',
    listTransferTickets: '/api/chat/transfer/list',
    resolveTransferTicket: '/api/chat/transfer/resolve'
  },
  knowledge: {
    overview: '/api/knowledge/overview',
    productList: '/api/knowledge/product/list',
    productCreate: '/api/knowledge/product/create',
    productUpdate: '/api/knowledge/product/update',
    productDelete: '/api/knowledge/product/delete',
    productDetail: '/api/knowledge/product/detail',
    productImport: '/api/knowledge/product/import',
    productExport: '/api/knowledge/product/export',
    activityList: '/api/knowledge/activity/list',
    activityCreate: '/api/knowledge/activity/create',
    activityUpdate: '/api/knowledge/activity/update',
    activityDelete: '/api/knowledge/activity/delete',
    activityDetail: '/api/knowledge/activity/detail',
    faqList: '/api/knowledge/faq/list',
    faqCreate: '/api/knowledge/faq/create',
    faqUpdate: '/api/knowledge/faq/update',
    faqDelete: '/api/knowledge/faq/delete',
    faqDetail: '/api/knowledge/faq/detail',
    faqImport: '/api/knowledge/faq/import',
    gapSummary: '/api/knowledge/gap/summary',
    gapList: '/api/knowledge/gap/list',
    gapResolve: '/api/knowledge/gap/resolve',
    gapVerify: '/api/knowledge/gap/verify',
    industryList: '/api/knowledge/industry/list',
    industryCreate: '/api/knowledge/industry/create',
    industryUpdate: '/api/knowledge/industry/update',
    industryDelete: '/api/knowledge/industry/delete',
    solutionList: '/api/knowledge/solution/list',
    solutionCreate: '/api/knowledge/solution/create',
    solutionUpdate: '/api/knowledge/solution/update',
    solutionDelete: '/api/knowledge/solution/delete',
    solutionDetail: '/api/knowledge/solution/detail',
    vectorize: '/api/knowledge/vectorize',
    vectorizeStatus: '/api/knowledge/vectorize/status'
  },
  scene: {
    runtimeOverview: '/api/scene/runtime/overview',
    intentList: '/api/scene/intent/list',
    intentCreate: '/api/scene/intent/create',
    intentUpdate: '/api/scene/intent/update',
    intentDelete: '/api/scene/intent/delete',
    intentDetail: '/api/scene/intent/detail',
    promptList: '/api/scene/prompt/list',
    promptCreate: '/api/scene/prompt/create',
    promptUpdate: '/api/scene/prompt/update',
    promptDelete: '/api/scene/prompt/delete',
    promptDetail: '/api/scene/prompt/detail',
    promptTest: '/api/scene/prompt/test',
    ruleList: '/api/scene/rule/list',
    ruleCreate: '/api/scene/rule/create',
    ruleUpdate: '/api/scene/rule/update',
    ruleDelete: '/api/scene/rule/delete',
    ruleDetail: '/api/scene/rule/detail',
    ruleEnable: '/api/scene/rule/enable',
    ruleDisable: '/api/scene/rule/disable',
    ruleTest: '/api/scene/rule/test'
  }
} as const
