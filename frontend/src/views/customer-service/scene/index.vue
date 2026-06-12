<script setup lang="ts">
/**
 * 场景配置 - 意图/提示词/规则管理
 */

import IntentConfig from './tabs/IntentConfig.vue'
import PromptConfig from './tabs/PromptConfig.vue'
import RuleConfig from './tabs/RuleConfig.vue'
import { fetchSceneRuntimeOverview } from '@/api/customer-service/scene'

// 当前激活的tab
const activeTab = ref('intent')
const overview = ref<any>({
  activeIntentCount: 0,
  activeSystemPromptCount: 0,
  activeRuleCount: 0,
  totalRuleHits: 0,
  totalPromptUsage: 0,
  topRules: [],
  topPrompts: []
})

// Tab配置
const tabs = [
  { name: 'intent', label: '意图管理', icon: '🎯', desc: '启用意图会按关键词参与运行时识别，LLM失败后兜底使用' },
  { name: 'prompt', label: '提示词模板', icon: '📝', desc: '启用的SYSTEM提示词会注入对话运行时，测试按钮可直接调LLM' },
  { name: 'rule', label: '场景规则', icon: '⚙️', desc: '启用规则会进入护栏策略，当前支持基础敏感过滤和转人工阈值' }
]

const loadOverview = async () => {
  try {
    overview.value = await fetchSceneRuntimeOverview()
  } catch {
    // 保持默认值
  }
}

onMounted(() => {
  loadOverview()
})
</script>

<template>
  <div class="scene-config">
    <!-- 头部 -->
    <div class="scene-header">
      <div class="header-title">
        <h2>⚙️ 场景配置</h2>
        <p>配置智能客服的意图识别、提示词模板和场景规则</p>
      </div>
    </div>

    <div class="runtime-overview">
      <div class="runtime-metric">
        <div class="runtime-metric__label">启用意图</div>
        <div class="runtime-metric__value">{{ overview.activeIntentCount }}</div>
      </div>
      <div class="runtime-metric">
        <div class="runtime-metric__label">SYSTEM Prompt</div>
        <div class="runtime-metric__value">{{ overview.activeSystemPromptCount }}</div>
      </div>
      <div class="runtime-metric">
        <div class="runtime-metric__label">启用规则</div>
        <div class="runtime-metric__value">{{ overview.activeRuleCount }}</div>
      </div>
      <div class="runtime-metric">
        <div class="runtime-metric__label">规则命中</div>
        <div class="runtime-metric__value">{{ overview.totalRuleHits }}</div>
      </div>
      <div class="runtime-metric">
        <div class="runtime-metric__label">Prompt 使用</div>
        <div class="runtime-metric__value">{{ overview.totalPromptUsage }}</div>
      </div>
      <div class="runtime-rank">
        <div class="runtime-rank__label">Top 规则</div>
        <div class="runtime-rank__content">
          {{ overview.topRules?.[0]?.name || '-' }}
        </div>
      </div>
      <div class="runtime-rank">
        <div class="runtime-rank__label">Top Prompt</div>
        <div class="runtime-rank__content">
          {{ overview.topPrompts?.[0]?.name || '-' }}
        </div>
      </div>
    </div>

    <!-- Tab内容 -->
    <div class="scene-content">
      <el-tabs v-model="activeTab" class="scene-tabs">
        <el-tab-pane
          v-for="tab in tabs"
          :key="tab.name"
          :name="tab.name"
        >
          <template #label>
            <el-tooltip :content="tab.desc" placement="top" effect="light" :show-after="500" :offset="8">
              <div class="tab-label">
                <span class="tab-icon">{{ tab.icon }}</span>
                <span class="tab-text">{{ tab.label }}</span>
              </div>
            </el-tooltip>
          </template>

          <template v-if="activeTab === 'intent'">
            <IntentConfig />
          </template>
          <template v-else-if="activeTab === 'prompt'">
            <PromptConfig />
          </template>
          <template v-else-if="activeTab === 'rule'">
            <RuleConfig />
          </template>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.scene-config {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.scene-header {
  margin-bottom: 12px;
  flex-shrink: 0;

  .header-title {
    h2 {
      margin: 0 0 8px 0;
      font-size: 22px;
      color: #333;
    }

    p {
      margin: 0;
      color: #666;
      font-size: 14px;
    }
  }
}

.runtime-overview {
  display: grid;
  grid-template-columns: repeat(5, minmax(96px, 1fr)) repeat(2, minmax(140px, 1.2fr));
  gap: 8px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.runtime-metric,
.runtime-rank {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  background-color: #fff;
}

.runtime-metric {
  &__label {
    margin-bottom: 4px;
    color: #999;
    font-size: 12px;
  }

  &__value {
    color: #333;
    font-size: 18px;
    font-weight: 600;
  }
}

.runtime-rank {
  &__label {
    margin-bottom: 4px;
    color: #999;
    font-size: 12px;
  }

  &__content {
    overflow: hidden;
    color: #333;
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.scene-content {
  flex: 1;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

@media (max-width: 1280px) {
  .runtime-overview {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.scene-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;

  :deep(.el-tabs__header) {
    margin: 0;
    padding: 0 20px;
    background-color: #fafafa;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.el-tabs__item) {
    height: 56px;
    padding: 0 24px;
  }

  :deep(.el-tabs__content) {
    flex: 1;
    overflow: hidden;
    padding: 0;
  }

  :deep(.el-tab-pane) {
    height: 100%;
    overflow-y: auto;
    padding: 20px;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;

  .tab-icon {
    font-size: 16px;
  }

  .tab-text {
    font-size: 14px;
  }

  .tab-badge {
    :deep(.el-badge__content) {
      font-size: 10px;
      height: 16px;
      line-height: 16px;
      padding: 0 5px;
    }
  }
}
</style>
