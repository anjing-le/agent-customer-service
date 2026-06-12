<script setup lang="ts">
/**
 * 对话中心 - 四栏布局
 * 1. 常见问题侧边栏
 * 2. 对话区域（iPhone模拟器样式）
 * 3. 知识选择区域
 * 4. Agent工作区（推理过程+知识召回）
 */

import QuestionSidebar from './components/QuestionSidebar.vue'
import ChatInterface from './components/ChatInterface.vue'
import KnowledgeSelect from './components/KnowledgeSelect.vue'
import AgentWorkArea from './components/AgentWorkArea.vue'
import { fetchChatRuntimeOverview } from '@/api/customer-service/chat'

// 选中的问题
const selectedQuestion = ref<string>('')
const selectedQuestionId = ref<string>('')

// 知识选择
const knowledgeSelection = ref<{ selectedProducts: number[]; selectedActivities: number[] }>({
  selectedProducts: [],
  selectedActivities: []
})

// 最新一次的 AI 响应（传给 AgentWorkArea 展示）
const lastAiResponse = ref<any>(null)
const overview = ref({
  totalSessions: 0,
  activeSessions: 0,
  todaySessions: 0,
  totalMessages: 0,
  todayMessages: 0,
  todayUserMessages: 0,
  todayAssistantMessages: 0,
  todayAgentReplies: 0,
  todaySafeReplies: 0,
  todayFallbackReplies: 0,
  averageMessagesPerSession: 0,
  recentSessions: [] as any[],
  recentAudits: [] as any[],
  qualitySummary: {
    totalAuditedReplies: 0,
    averageConfidence: 0,
    fallbackRate: 0,
    unsafeRate: 0,
    averageKnowledgeEvidenceCount: 0,
    averageRuleHitCount: 0,
    averagePromptRenderCount: 0
  },
  dailyTrends: [] as any[]
})

const loadOverview = async () => {
  overview.value = await fetchChatRuntimeOverview()
}

// 处理问题选择
const handleQuestionSelect = (question: { id: string; text: string }) => {
  selectedQuestion.value = question.text
  selectedQuestionId.value = question.id
}

// 问题发送后清空选中
const handleQuestionSent = () => {
  selectedQuestion.value = ''
  selectedQuestionId.value = ''
  loadOverview()
}

// 知识选择变化
const handleKnowledgeChange = (data: { selectedProducts: number[]; selectedActivities: number[] }) => {
  knowledgeSelection.value = data
}

// AI 响应更新
const handleAiResponse = (response: any) => {
  lastAiResponse.value = response
  loadOverview()
}

onMounted(() => {
  loadOverview()
})
</script>

<template>
  <div class="chat-page">
    <div class="runtime-overview">
      <div class="runtime-metric">
        <span>总会话</span>
        <strong>{{ overview.totalSessions }}</strong>
      </div>
      <div class="runtime-metric">
        <span>活跃会话</span>
        <strong>{{ overview.activeSessions }}</strong>
      </div>
      <div class="runtime-metric">
        <span>今日新增</span>
        <strong>{{ overview.todaySessions }}</strong>
      </div>
      <div class="runtime-metric">
        <span>今日消息</span>
        <strong>{{ overview.todayMessages }}</strong>
      </div>
      <div class="runtime-metric">
        <span>Agent 回复</span>
        <strong>{{ overview.todayAgentReplies }}</strong>
      </div>
      <div class="runtime-metric">
        <span>安全 / 兜底</span>
        <strong>{{ overview.todaySafeReplies }} / {{ overview.todayFallbackReplies }}</strong>
      </div>
      <div class="runtime-recent">
        <span>最近审计</span>
        <strong>
          {{ overview.recentAudits?.[0]?.intentName || overview.recentSessions?.[0]?.lastMessage || '暂无记录' }}
        </strong>
      </div>
      <div class="runtime-quality">
        <div>
          <span>平均置信度</span>
          <strong>{{ overview.qualitySummary?.averageConfidence }}</strong>
        </div>
        <div>
          <span>兜底率</span>
          <strong>{{ overview.qualitySummary?.fallbackRate }}%</strong>
        </div>
        <div>
          <span>不安全率</span>
          <strong>{{ overview.qualitySummary?.unsafeRate }}%</strong>
        </div>
      </div>
      <div class="runtime-trend">
        <div
          v-for="item in overview.dailyTrends"
          :key="item.date"
          class="trend-item"
        >
          <span>{{ item.date.slice(5) }}</span>
          <strong>{{ item.replies }}</strong>
          <em>{{ item.fallbackReplies }}</em>
        </div>
      </div>
    </div>

    <div class="chat-center">
      <!-- 1. 常见问题侧边栏 -->
      <div class="chat-center__sidebar">
        <QuestionSidebar
          :selected-question="selectedQuestion"
          @select="handleQuestionSelect"
        />
      </div>

      <!-- 2. 对话区域 - iPhone模拟器样式 -->
      <div class="chat-center__chat">
        <ChatInterface
          :selected-question="selectedQuestion"
          :selected-question-id="selectedQuestionId"
          :knowledge-selection="knowledgeSelection"
          @question-sent="handleQuestionSent"
          @ai-response="handleAiResponse"
        />
      </div>

      <!-- 3. 知识选择区域 -->
      <div class="chat-center__knowledge">
        <KnowledgeSelect @selection-change="handleKnowledgeChange" />
      </div>

      <!-- 4. Agent工作区 -->
      <div class="chat-center__agent">
        <AgentWorkArea :ai-response="lastAiResponse" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  background-color: #f7f5ee;
  min-height: calc(100vh - 120px);
  box-sizing: border-box;
}

.runtime-overview {
  display: grid;
  grid-template-columns: repeat(6, minmax(118px, 1fr)) minmax(220px, 1.6fr);
  gap: 8px;
}

.runtime-metric,
.runtime-recent,
.runtime-quality,
.runtime-trend {
  min-height: 58px;
  padding: 10px 12px;
  border: 1px solid #dedbd1;
  border-radius: 8px;
  background: #fffdf7;
  box-sizing: border-box;

  span {
    display: block;
    color: #7b7567;
    font-size: 12px;
    line-height: 18px;
  }

  strong {
    display: block;
    margin-top: 2px;
    color: #2d2a24;
    font-size: 20px;
    line-height: 26px;
    font-weight: 700;
  }
}

.runtime-recent strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}

.runtime-quality {
  grid-column: span 3;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;

  span {
    display: block;
    color: #7b7567;
    font-size: 12px;
    line-height: 18px;
  }

  strong {
    color: #2d2a24;
    font-size: 18px;
    line-height: 24px;
  }
}

.runtime-trend {
  grid-column: span 4;
  display: grid;
  grid-template-columns: repeat(7, minmax(46px, 1fr));
  gap: 6px;
}

.trend-item {
  min-width: 0;
  padding: 4px 6px;
  border-radius: 6px;
  background: #f3f0e7;
  text-align: center;

  span,
  em {
    display: block;
    color: #7b7567;
    font-size: 11px;
    line-height: 15px;
    font-style: normal;
  }

  strong {
    display: block;
    color: #2d2a24;
    font-size: 16px;
    line-height: 20px;
  }
}

.chat-center {
  display: flex;
  gap: 12px;
  height: calc(100vh - 198px);
  width: 100%;
  box-sizing: border-box;

  &__sidebar {
    width: 18%;
    min-width: 200px;
    height: 100%;
    flex-shrink: 0;
  }

  &__chat {
    width: 22%;
    min-width: 340px;
    height: 100%;
    background-color: #f0f1f4;
    border: 6px solid rgb(204, 233, 255);
    border-radius: 40px;
    overflow: hidden;
    flex-shrink: 0;
  }

  &__knowledge {
    width: 20%;
    min-width: 200px;
    height: 100%;
    flex-shrink: 0;
  }

  &__agent {
    flex: 1;
    min-width: 300px;
    height: 100%;
  }
}

// 响应式适配
@media (max-width: 1400px) {
  .runtime-overview {
    grid-template-columns: repeat(3, minmax(140px, 1fr));
  }

  .runtime-quality,
  .runtime-trend {
    grid-column: span 3;
  }

  .chat-center {
    &__sidebar {
      width: 16%;
      min-width: 180px;
    }

    &__chat {
      width: 24%;
    }

    &__knowledge {
      width: 18%;
      min-width: 180px;
    }
  }
}

@media (max-width: 1200px) {
  .chat-page {
    min-height: auto;
  }

  .runtime-overview {
    grid-template-columns: repeat(2, minmax(140px, 1fr));
  }

  .runtime-quality,
  .runtime-trend {
    grid-column: span 2;
  }

  .chat-center {
    flex-wrap: wrap;
    height: auto;

    &__sidebar {
      width: 100%;
      height: auto;
      max-height: 150px;
    }

    &__chat {
      width: 35%;
      height: calc(100vh - 300px);
    }

    &__knowledge {
      width: 30%;
      height: calc(100vh - 300px);
    }

    &__agent {
      width: 30%;
      height: calc(100vh - 300px);
    }
  }
}
</style>
