<script setup lang="ts">
/**
 * Agent工作区组件
 * 展示用户画像、服务流程、知识召回、推理过程
 * 数据来源：后端 ChatService.sendMessage 返回的真实处理结果
 */

interface Props {
  aiResponse?: any
}

const props = defineProps<Props>()

const activeTab = ref('reasoning')

const userProfile = ref({
  userId: 'U202501020001',
  nickname: '演示用户',
  level: 'VIP会员',
  tags: ['高频购买', '价格敏感', '服装偏好'],
  recentOrders: 5,
  totalAmount: 2680,
  lastActive: '刚刚'
})

const serviceFlow = ref([
  { step: 1, name: '场景识别', status: 'pending', result: '等待输入' },
  { step: 2, name: '意图分析', status: 'pending', result: '等待输入' },
  { step: 3, name: '情绪判断', status: 'pending', result: '等待输入' },
  { step: 4, name: '知识检索', status: 'pending', result: '等待输入' },
  { step: 5, name: '生成回复', status: 'pending', result: '等待输入' }
])

const knowledgeRecall = ref<any>({ products: [], faqs: [], activities: [] })

const reasoningProcess = ref<any[]>([])
const reliability = ref<any>({
  replyEngine: '-',
  safe: true,
  fallbackRequired: false,
  fallbackReason: '',
  userVisibleNotice: '',
  policyTags: [],
  ruleHits: [],
  promptRenders: []
})
const sessionQuality = ref<any>({
  totalAuditedReplies: 0,
  fallbackReplies: 0,
  unsafeReplies: 0,
  lowConfidenceReplies: 0,
  averageConfidence: 0,
  fallbackRate: 0,
  unsafeRate: 0,
  reliabilityScore: 100,
  riskLevel: 'LOW',
  primaryFallbackReason: ''
})
const sessionAudits = ref<any[]>([])

const riskTagType = computed(() => {
  if (sessionQuality.value.riskLevel === 'HIGH') return 'danger'
  if (sessionQuality.value.riskLevel === 'MEDIUM') return 'warning'
  return 'success'
})

const formatTime = (ts: string | null) => {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`
  } catch {
    return ''
  }
}

watch(() => props.aiResponse, (res) => {
  if (!res) return

  console.log('[AgentWorkArea] 收到AI响应:', JSON.stringify(res).substring(0, 500))

  const totalKnowledge = (res.knowledgeRecall?.products?.length || 0)
    + (res.knowledgeRecall?.faqs?.length || 0)
    + (res.knowledgeRecall?.activities?.length || 0)

  serviceFlow.value = [
    { step: 1, name: '场景识别', status: 'completed', result: res.sceneType || '通用咨询' },
    { step: 2, name: '意图分析', status: 'completed', result: res.intent?.intentName || '通用咨询' },
    { step: 3, name: '情绪判断', status: 'completed', result: res.emotion || '中性' },
    { step: 4, name: '知识检索', status: 'completed', result: `检索到${totalKnowledge}条相关知识` },
    { step: 5, name: '生成回复', status: 'completed', result: '已完成' }
  ]

  knowledgeRecall.value = {
    products: (res.knowledgeRecall?.products || []).map((p: any) => ({
      id: p.productId,
      name: p.productName,
      score: p.score ?? 0.8,
      source: p.source || '知识库'
    })),
    faqs: (res.knowledgeRecall?.faqs || []).map((f: any) => ({
      id: f.faqId,
      question: f.question,
      answer: f.answer,
      score: f.score ?? 0.8
    })),
    activities: (res.knowledgeRecall?.activities || []).map((a: any) => ({
      id: a.activityId,
      name: a.activityName,
      score: a.score ?? 0.8
    }))
  }

  reasoningProcess.value = (res.reasoningProcess || []).map((r: any) => ({
    step: r.step,
    title: r.title,
    content: r.content,
    timestamp: formatTime(r.timestamp)
  }))

  reliability.value = {
    replyEngine: res.reliability?.replyEngine || '-',
    safe: res.reliability?.safe ?? true,
    fallbackRequired: res.reliability?.fallbackRequired ?? false,
    fallbackReason: res.reliability?.fallbackReason || '',
    userVisibleNotice: res.reliability?.userVisibleNotice || '',
    policyTags: res.reliability?.policyTags || [],
    ruleHits: res.reliability?.ruleHits || [],
    promptRenders: res.reliability?.promptRenders || []
  }
  sessionQuality.value = {
    totalAuditedReplies: res.sessionQuality?.totalAuditedReplies || 0,
    fallbackReplies: res.sessionQuality?.fallbackReplies || 0,
    unsafeReplies: res.sessionQuality?.unsafeReplies || 0,
    lowConfidenceReplies: res.sessionQuality?.lowConfidenceReplies || 0,
    averageConfidence: res.sessionQuality?.averageConfidence || 0,
    fallbackRate: res.sessionQuality?.fallbackRate || 0,
    unsafeRate: res.sessionQuality?.unsafeRate || 0,
    reliabilityScore: res.sessionQuality?.reliabilityScore ?? 100,
    riskLevel: res.sessionQuality?.riskLevel || 'LOW',
    primaryFallbackReason: res.sessionQuality?.primaryFallbackReason || ''
  }
  sessionAudits.value = res.sessionAudits || []

  userProfile.value.lastActive = '刚刚'
})
</script>

<template>
  <div class="agent-work-area">
    <!-- 用户画像区域 -->
    <div class="user-profile-section">
      <div class="section-header">
        <span class="section-title">👤 用户画像</span>
      </div>
      <div class="user-profile-content">
        <div class="profile-main">
          <div class="profile-avatar">👤</div>
          <div class="profile-info">
            <div class="profile-nickname">{{ userProfile.nickname }}</div>
            <el-tag size="small" type="warning">{{ userProfile.level }}</el-tag>
          </div>
        </div>
        <div class="profile-stats">
          <div class="stat-item">
            <div class="stat-value">{{ userProfile.recentOrders }}</div>
            <div class="stat-label">近期订单</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">¥{{ userProfile.totalAmount }}</div>
            <div class="stat-label">累计消费</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ userProfile.lastActive }}</div>
            <div class="stat-label">最近活跃</div>
          </div>
        </div>
        <div class="profile-tags">
          <el-tag v-for="tag in userProfile.tags" :key="tag" size="small">
            {{ tag }}
          </el-tag>
        </div>
      </div>
    </div>

    <!-- 服务流程区域 -->
    <div class="service-flow-section">
      <div class="section-header">
        <span class="section-title">🔄 服务流程</span>
      </div>
      <div class="flow-steps">
        <div
          v-for="item in serviceFlow"
          :key="item.step"
          class="flow-step"
          :class="[`flow-step--${item.status}`]"
        >
          <div class="flow-step__indicator">
            <span v-if="item.status === 'completed'">✓</span>
            <span v-else-if="item.status === 'active'" class="loading-spinner"></span>
            <span v-else>{{ item.step }}</span>
          </div>
          <div class="flow-step__content">
            <div class="flow-step__name">{{ item.name }}</div>
            <div class="flow-step__result">{{ item.result }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab切换区域 -->
    <div class="detail-section">
      <el-tabs v-model="activeTab" class="detail-tabs">
        <el-tab-pane label="推理过程" name="reasoning">
          <div v-if="reasoningProcess.length === 0" class="empty-hint">
            💡 发送消息后，这里将展示 AI 的推理过程
          </div>
          <div v-else class="reasoning-list">
            <div
              v-for="item in reasoningProcess"
              :key="item.step"
              class="reasoning-item"
            >
              <div class="reasoning-item__step">{{ item.step }}</div>
              <div class="reasoning-item__content">
                <div class="reasoning-item__title">
                  {{ item.title }}
                  <span class="reasoning-item__time">{{ item.timestamp }}</span>
                </div>
                <div class="reasoning-item__desc">{{ item.content }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="知识召回" name="knowledge">
          <div v-if="knowledgeRecall.products.length === 0 && knowledgeRecall.faqs.length === 0 && knowledgeRecall.activities.length === 0" class="empty-hint">
            💡 发送消息后，这里将展示知识召回结果
          </div>
          <div v-else class="knowledge-recall">
            <!-- 商品召回 -->
            <div v-if="knowledgeRecall.products.length > 0" class="recall-group">
              <div class="recall-group__title">🛍️ 商品召回</div>
              <div
                v-for="item in knowledgeRecall.products"
                :key="item.id"
                class="recall-item"
              >
                <div class="recall-item__name">{{ item.name }}</div>
                <div class="recall-item__meta">
                  <span class="recall-item__source">{{ item.source }}</span>
                  <span class="recall-item__score">{{ (item.score * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>

            <!-- FAQ召回 -->
            <div v-if="knowledgeRecall.faqs.length > 0" class="recall-group">
              <div class="recall-group__title">❓ FAQ召回</div>
              <div
                v-for="item in knowledgeRecall.faqs"
                :key="item.id"
                class="recall-item"
              >
                <div class="recall-item__name">{{ item.question }}</div>
                <div class="recall-item__answer">{{ item.answer }}</div>
                <div class="recall-item__meta">
                  <span class="recall-item__score">匹配度：{{ (item.score * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>

            <!-- 活动召回 -->
            <div v-if="knowledgeRecall.activities.length > 0" class="recall-group">
              <div class="recall-group__title">🎁 活动召回</div>
              <div
                v-for="item in knowledgeRecall.activities"
                :key="item.id"
                class="recall-item"
              >
                <div class="recall-item__name">{{ item.name }}</div>
                <div class="recall-item__meta">
                  <span class="recall-item__score">{{ (item.score * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="可靠性" name="reliability">
          <div v-if="!props.aiResponse" class="empty-hint">
            发送消息后，这里将展示回复引擎、兜底原因和护栏标签
          </div>
          <div v-else class="reliability-panel">
            <div class="reliability-grid">
              <div class="reliability-metric">
                <div class="reliability-metric__label">回复引擎</div>
                <div class="reliability-metric__value">{{ reliability.replyEngine }}</div>
              </div>
              <div class="reliability-metric">
                <div class="reliability-metric__label">安全状态</div>
                <div class="reliability-metric__value">
                  <el-tag :type="reliability.safe ? 'success' : 'danger'" size="small">
                    {{ reliability.safe ? '通过' : '拦截' }}
                  </el-tag>
                </div>
              </div>
              <div class="reliability-metric">
                <div class="reliability-metric__label">兜底状态</div>
                <div class="reliability-metric__value">
                  <el-tag :type="reliability.fallbackRequired ? 'warning' : 'success'" size="small">
                    {{ reliability.fallbackRequired ? '已兜底' : '未兜底' }}
                  </el-tag>
                </div>
              </div>
            </div>

            <div class="session-quality">
              <div class="session-quality__header">
                <div>
                  <div class="session-quality__title">会话质检摘要</div>
                  <div class="session-quality__desc">基于当前会话已审计回复聚合</div>
                </div>
                <el-tag :type="riskTagType" size="small">{{ sessionQuality.riskLevel }}</el-tag>
              </div>
              <div class="session-quality__score">
                <span>可靠性评分</span>
                <strong>{{ sessionQuality.reliabilityScore }}</strong>
              </div>
              <div class="session-quality__grid">
                <div>
                  <span>已审计</span>
                  <strong>{{ sessionQuality.totalAuditedReplies }}</strong>
                </div>
                <div>
                  <span>平均置信度</span>
                  <strong>{{ sessionQuality.averageConfidence }}</strong>
                </div>
                <div>
                  <span>兜底率</span>
                  <strong>{{ sessionQuality.fallbackRate }}%</strong>
                </div>
                <div>
                  <span>不安全率</span>
                  <strong>{{ sessionQuality.unsafeRate }}%</strong>
                </div>
              </div>
              <div class="session-quality__reasons">
                <span>主要兜底原因</span>
                <strong>{{ sessionQuality.primaryFallbackReason || '暂无' }}</strong>
              </div>
            </div>

            <div class="reliability-block">
              <div class="reliability-block__label">质检明细</div>
              <div v-if="sessionAudits.length" class="audit-detail-list">
                <div
                  v-for="audit in sessionAudits"
                  :key="audit.messageId"
                  class="audit-detail-item"
                >
                  <div class="audit-detail-item__header">
                    <span>{{ audit.intentName || '通用咨询' }}</span>
                    <el-tag
                      :type="audit.safe === false ? 'danger' : audit.fallbackRequired ? 'warning' : 'success'"
                      size="small"
                    >
                      {{ audit.safe === false ? '拦截' : audit.fallbackRequired ? '兜底' : '通过' }}
                    </el-tag>
                  </div>
                  <div class="audit-detail-item__meta">
                    <span>{{ audit.replyEngine || '-' }}</span>
                    <span>置信度 {{ audit.confidence ?? 0 }}</span>
                    <span>召回 {{ audit.knowledgeEvidenceCount || 0 }}</span>
                    <span>规则 {{ audit.ruleHitCount || 0 }}</span>
                    <span>Prompt {{ audit.promptRenderCount || 0 }}</span>
                  </div>
                  <div v-if="audit.fallbackReason" class="audit-detail-item__reason">
                    {{ audit.fallbackReason }}
                  </div>
                </div>
              </div>
              <div v-else class="reliability-block__content reliability-block__content--muted">
                暂无会话审计明细
              </div>
            </div>

            <div v-if="reliability.fallbackReason" class="reliability-block">
              <div class="reliability-block__label">兜底原因</div>
              <div class="reliability-block__content">{{ reliability.fallbackReason }}</div>
            </div>

            <div v-if="reliability.userVisibleNotice" class="reliability-block">
              <div class="reliability-block__label">用户提示</div>
              <div class="reliability-block__content">{{ reliability.userVisibleNotice }}</div>
            </div>

            <div class="reliability-block">
              <div class="reliability-block__label">策略标签</div>
              <div v-if="reliability.policyTags.length" class="policy-tags">
                <el-tag
                  v-for="tag in reliability.policyTags"
                  :key="tag"
                  size="small"
                  type="info"
                >
                  {{ tag }}
                </el-tag>
              </div>
              <div v-else class="reliability-block__content reliability-block__content--muted">
                暂无策略标签
              </div>
            </div>

            <div class="reliability-block">
              <div class="reliability-block__label">命中规则</div>
              <div v-if="reliability.ruleHits.length" class="rule-hit-list">
                <div
                  v-for="rule in reliability.ruleHits"
                  :key="rule.ruleCode"
                  class="rule-hit-item"
                >
                  <div class="rule-hit-item__header">
                    <span class="rule-hit-item__name">{{ rule.ruleName }}</span>
                    <code class="rule-hit-item__code">{{ rule.ruleCode }}</code>
                  </div>
                  <div class="rule-hit-item__meta">
                    <span>P{{ rule.priority }}</span>
                    <span>{{ rule.ruleType }}</span>
                    <span>{{ rule.action }}</span>
                    <span>{{ rule.conditionSource || 'BUILT_IN' }}</span>
                  </div>
                  <div class="rule-hit-item__reason">{{ rule.reason }}</div>
                </div>
              </div>
              <div v-else class="reliability-block__content reliability-block__content--muted">
                暂无命中规则
              </div>
            </div>

            <div class="reliability-block">
              <div class="reliability-block__label">渲染提示词</div>
              <div v-if="reliability.promptRenders.length" class="prompt-render-list">
                <div
                  v-for="prompt in reliability.promptRenders"
                  :key="prompt.promptCode"
                  class="prompt-render-item"
                >
                  <div class="prompt-render-item__header">
                    <span class="prompt-render-item__name">{{ prompt.promptName }}</span>
                    <code class="prompt-render-item__code">{{ prompt.promptCode }}</code>
                  </div>
                  <div class="prompt-render-item__meta">
                    <span>{{ prompt.promptType }}</span>
                    <span>{{ prompt.sceneType || '全局' }}</span>
                  </div>
                  <pre class="prompt-render-item__content">{{ prompt.renderedContent }}</pre>
                </div>
              </div>
              <div v-else class="reliability-block__content reliability-block__content--muted">
                暂无渲染提示词
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.agent-work-area {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}

.section-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

// 用户画像区域
.user-profile-section {
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
}

.user-profile-content {
  padding: 16px;
}

.profile-main {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.profile-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: #e6f7ff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-right: 12px;
}

.profile-nickname {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.profile-stats {
  display: flex;
  justify-content: space-between;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.stat-item {
  text-align: center;

  .stat-value {
    font-size: 16px;
    font-weight: 600;
    color: #1890ff;
  }

  .stat-label {
    font-size: 11px;
    color: #999;
    margin-top: 2px;
  }
}

.profile-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

// 服务流程区域
.service-flow-section {
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
}

.flow-steps {
  display: flex;
  padding: 16px;
  overflow-x: auto;
  gap: 8px;
}

.flow-step {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background-color: #f5f5f5;
  border-radius: 8px;
  min-width: 120px;
  flex-shrink: 0;

  &--completed {
    background-color: #f6ffed;

    .flow-step__indicator {
      background-color: #52c41a;
      color: #fff;
    }
  }

  &--active {
    background-color: #e6f7ff;

    .flow-step__indicator {
      background-color: #1890ff;
      color: #fff;
    }
  }

  &__indicator {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background-color: #d9d9d9;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
    margin-right: 8px;
    flex-shrink: 0;
  }

  &__content {
    min-width: 0;
  }

  &__name {
    font-size: 12px;
    font-weight: 500;
    color: #333;
  }

  &__result {
    font-size: 11px;
    color: #666;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.loading-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

// Tab详情区域
.detail-section {
  flex: 1;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;

  :deep(.el-tabs__header) {
    margin: 0;
    padding: 0 16px;
    flex-shrink: 0;
  }

  :deep(.el-tabs__content) {
    flex: 1;
    overflow: hidden;
  }

  :deep(.el-tab-pane) {
    height: 100%;
    overflow-y: auto;
    padding: 16px;
  }
}

.empty-hint {
  padding: 24px 16px;
  text-align: center;
  color: #999;
  font-size: 13px;
}

// 推理过程
.reasoning-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reasoning-item {
  display: flex;
  align-items: flex-start;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 8px;

  &__step {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background-color: #1890ff;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
    margin-right: 12px;
    flex-shrink: 0;
  }

  &__content {
    flex: 1;
    min-width: 0;
  }

  &__title {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    font-weight: 500;
    color: #333;
    margin-bottom: 4px;
  }

  &__time {
    font-size: 11px;
    color: #999;
    font-weight: 400;
  }

  &__desc {
    font-size: 12px;
    color: #666;
    line-height: 1.5;
  }
}

// 知识召回
.knowledge-recall {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.recall-group {
  &__title {
    font-size: 13px;
    font-weight: 600;
    color: #333;
    margin-bottom: 8px;
  }
}

.recall-item {
  padding: 10px 12px;
  background-color: #fafafa;
  border-radius: 6px;
  margin-bottom: 8px;

  &__name {
    font-size: 13px;
    font-weight: 500;
    color: #333;
    margin-bottom: 4px;
  }

  &__answer {
    font-size: 12px;
    color: #666;
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__meta {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    color: #999;
  }

  &__score {
    color: #1890ff;
    font-weight: 500;
  }
}

.reliability-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reliability-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.reliability-metric {
  padding: 10px;
  background-color: #fafafa;
  border-radius: 6px;

  &__label {
    font-size: 11px;
    color: #999;
    margin-bottom: 6px;
  }

  &__value {
    min-height: 22px;
    font-size: 13px;
    font-weight: 600;
    color: #333;
    word-break: break-word;
  }
}

.reliability-block {
  padding: 12px;
  background-color: #fafafa;
  border-radius: 6px;

  &__label {
    font-size: 12px;
    font-weight: 600;
    color: #333;
    margin-bottom: 6px;
  }

  &__content {
    font-size: 12px;
    color: #666;
    line-height: 1.5;
    word-break: break-word;

    &--muted {
      color: #999;
    }
  }
}

.session-quality {
  padding: 12px;
  border: 1px solid #e8eef7;
  border-radius: 6px;
  background-color: #f8fbff;

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 10px;
  }

  &__title {
    color: #333;
    font-size: 13px;
    font-weight: 600;
    line-height: 18px;
  }

  &__desc {
    color: #8a94a6;
    font-size: 11px;
    line-height: 16px;
  }

  &__score {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;

    span {
      color: #606a7a;
      font-size: 12px;
    }

    strong {
      color: #1f5fbf;
      font-size: 24px;
      line-height: 28px;
      font-weight: 700;
    }
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 6px;

    div {
      min-width: 0;
      padding: 8px;
      border-radius: 6px;
      background-color: #fff;
    }

    span,
    strong {
      display: block;
    }

    span {
      color: #8a94a6;
      font-size: 11px;
      line-height: 16px;
    }

    strong {
      color: #2f3747;
      font-size: 13px;
      line-height: 18px;
    }
  }

  &__reasons {
    margin-top: 8px;
    padding: 8px;
    border-radius: 6px;
    background-color: #fff;

    span,
    strong {
      display: block;
    }

    span {
      color: #8a94a6;
      font-size: 11px;
      line-height: 16px;
    }

    strong {
      color: #2f3747;
      font-size: 12px;
      line-height: 18px;
      word-break: break-word;
    }
  }
}

.policy-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.rule-hit-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.audit-detail-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 260px;
  overflow-y: auto;
}

.prompt-render-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.audit-detail-item {
  padding: 10px;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  background-color: #fff;

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 6px;

    span {
      min-width: 0;
      overflow: hidden;
      color: #333;
      font-size: 13px;
      font-weight: 600;
      line-height: 18px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  &__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 8px;
    color: #999;
    font-size: 11px;
    line-height: 16px;
  }

  &__reason {
    margin-top: 6px;
    color: #666;
    font-size: 12px;
    line-height: 18px;
    word-break: break-word;
  }
}

.rule-hit-item {
  padding: 10px;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 6px;

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 6px;
  }

  &__name {
    min-width: 0;
    font-size: 13px;
    font-weight: 600;
    color: #333;
  }

  &__code {
    flex-shrink: 0;
    padding: 1px 6px;
    background-color: #f5f5f5;
    border-radius: 4px;
    color: #666;
    font-size: 11px;
  }

  &__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 6px;
    color: #999;
    font-size: 11px;
  }

  &__reason {
    color: #666;
    font-size: 12px;
    line-height: 1.5;
  }
}

.prompt-render-item {
  padding: 10px;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 6px;

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 6px;
  }

  &__name {
    min-width: 0;
    font-size: 13px;
    font-weight: 600;
    color: #333;
  }

  &__code {
    flex-shrink: 0;
    padding: 1px 6px;
    background-color: #f5f5f5;
    border-radius: 4px;
    color: #666;
    font-size: 11px;
  }

  &__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 8px;
    color: #999;
    font-size: 11px;
  }

  &__content {
    max-height: 160px;
    margin: 0;
    padding: 8px;
    overflow: auto;
    border-radius: 4px;
    background-color: #fafafa;
    color: #555;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
  }
}
</style>
