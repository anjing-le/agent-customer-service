<script setup lang="ts">
/**
 * 场景规则管理 - 调用后端 SceneController 接口
 */
import {
  fetchListRules,
  fetchCreateRule,
  fetchUpdateRule,
  fetchDeleteRule,
  fetchEnableRule,
  fetchDisableRule,
  fetchTestRule
} from '@/api/customer-service/scene'

interface Rule {
  id: number
  ruleName: string
  ruleCode: string
  ruleType: string
  description: string
  conditions: string
  actions: string
  triggerCount: number
  priority: number
  enabled: boolean
  effectiveTime: string
  expireTime: string
  updatedAt: string
}

const tableData = ref<Rule[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const testDialogVisible = ref(false)
const currentRule = ref<Rule | null>(null)
const submitting = ref(false)
const ruleFormRef = ref()
const ruleForm = reactive({
  id: undefined as number | undefined,
  ruleName: '',
  ruleCode: '',
  ruleType: '转接规则',
  description: '',
  conditions: '',
  actions: '',
  priority: 5,
  effectiveTime: '',
  expireTime: '',
  enabled: true
})
const testForm = reactive({
  userMessage: '我想退货，但是不确定要怎么操作',
  sceneType: '售后服务',
  intentCode: 'RETURN_EXCHANGE',
  intentName: '退换货咨询',
  confidence: 0.62,
  emotion: '中性',
  knowledgeCount: 1,
  hasReliableKnowledge: true,
  contextText: '{}',
  result: null as any
})

const conditionExample = `{
  "all": [
    { "field": "intentCode", "op": "eq", "value": "RETURN_EXCHANGE" },
    { "field": "confidence", "op": "lt", "value": 0.7 }
  ]
}`

const actionExample = `{
  "reason": "退换货意图置信度不足，需要澄清",
  "action": "TRANSFER_OR_CLARIFY"
}`

const loadData = async () => {
  loading.value = true
  try {
    const res = await fetchListRules({})
    tableData.value = res.records || []
  } catch {
    // 后端不可用时保持空数据
  } finally {
    loading.value = false
  }
}

onMounted(() => { loadData() })

const handleAdd = () => {
  currentRule.value = null
  resetRuleForm()
  dialogVisible.value = true
}

const handleEdit = (row: Rule) => {
  currentRule.value = row
  Object.assign(ruleForm, {
    id: row.id,
    ruleName: row.ruleName || '',
    ruleCode: row.ruleCode || '',
    ruleType: row.ruleType || '转接规则',
    description: row.description || '',
    conditions: row.conditions || '',
    actions: row.actions || '',
    priority: row.priority ?? 5,
    effectiveTime: row.effectiveTime || '',
    expireTime: row.expireTime || '',
    enabled: row.enabled ?? true
  })
  dialogVisible.value = true
}

const resetRuleForm = () => {
  Object.assign(ruleForm, {
    id: undefined,
    ruleName: '',
    ruleCode: '',
    ruleType: '转接规则',
    description: '',
    conditions: '',
    actions: '',
    priority: 5,
    effectiveTime: '',
    expireTime: '',
    enabled: true
  })
  ruleFormRef.value?.clearValidate?.()
}

const parseJson = (value: string, label: string) => {
  if (!value?.trim()) return null
  try {
    return JSON.parse(value)
  } catch {
    throw new Error(`${label}不是合法 JSON`)
  }
}

const validateConditions = (_rule: any, value: string, callback: (error?: Error) => void) => {
  try {
    const json = parseJson(value, '规则条件')
    if (json && !json.all && !json.any && !json.field) {
      callback(new Error('规则条件必须包含 all、any 或 field'))
      return
    }
    callback()
  } catch (error: any) {
    callback(error)
  }
}

const validateActions = (_rule: any, value: string, callback: (error?: Error) => void) => {
  try {
    const json = parseJson(value, '规则动作')
    if (json && (Array.isArray(json) || typeof json !== 'object')) {
      callback(new Error('规则动作必须是 JSON 对象'))
      return
    }
    callback()
  } catch (error: any) {
    callback(error)
  }
}

const ruleFormRules = {
  ruleName: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  ruleCode: [{ required: true, message: '请输入规则编码', trigger: 'blur' }],
  ruleType: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  conditions: [{ validator: validateConditions, trigger: 'blur' }],
  actions: [{ validator: validateActions, trigger: 'blur' }]
}

const formatJsonField = (field: 'conditions' | 'actions') => {
  try {
    const json = parseJson(ruleForm[field], field === 'conditions' ? '规则条件' : '规则动作')
    if (json) ruleForm[field] = JSON.stringify(json, null, 2)
  } catch (error: any) {
    ElMessage.warning(error.message)
  }
}

const fillExample = () => {
  ruleForm.conditions = conditionExample
  ruleForm.actions = actionExample
}

const handleSubmit = async () => {
  await ruleFormRef.value?.validate?.()
  submitting.value = true
  try {
    const payload = { ...ruleForm }
    if (payload.id) {
      await fetchUpdateRule(payload)
    } else {
      await fetchCreateRule(payload)
    }
    ElMessage.success(payload.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleTest = (row: Rule) => {
  currentRule.value = row
  Object.assign(testForm, {
    userMessage: '我想退货，但是不确定要怎么操作',
    sceneType: '售后服务',
    intentCode: 'RETURN_EXCHANGE',
    intentName: '退换货咨询',
    confidence: 0.62,
    emotion: '中性',
    knowledgeCount: 1,
    hasReliableKnowledge: true,
    contextText: '{}',
    result: null
  })
  testDialogVisible.value = true
}

const handleRunTest = async () => {
  if (!currentRule.value) return
  try {
    const context = testForm.contextText.trim() ? JSON.parse(testForm.contextText) : {}
    testForm.result = await fetchTestRule({
      ruleId: currentRule.value.id,
      userMessage: testForm.userMessage,
      sceneType: testForm.sceneType,
      intentCode: testForm.intentCode,
      intentName: testForm.intentName,
      confidence: testForm.confidence,
      emotion: testForm.emotion,
      knowledgeCount: testForm.knowledgeCount,
      hasReliableKnowledge: testForm.hasReliableKnowledge,
      context
    })
  } catch {
    ElMessage.warning('上下文必须是合法 JSON')
  }
}

const handleDelete = (row: Rule) => {
  ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
    type: 'warning'
  }).then(async () => {
    await fetchDeleteRule(row.id)
    ElMessage.success('删除成功')
    loadData()
  }).catch(() => {})
}

const handleToggleEnabled = async (row: Rule) => {
  try {
    if (row.enabled) {
      await fetchDisableRule(row.id)
    } else {
      await fetchEnableRule(row.id)
    }
    row.enabled = !row.enabled
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch {
    ElMessage.error('操作失败')
  }
}

type RuleTagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getRuleTypeColor = (type: string): RuleTagType => {
  const colorMap: Record<string, RuleTagType> = {
    '路由规则': 'primary',
    '安全规则': 'danger',
    '转接规则': 'warning',
    '时间规则': 'info'
  }
  return colorMap[type] || 'info'
}
</script>

<template>
  <div class="rule-config">
    <!-- 操作栏 -->
    <div class="action-bar">
      <el-input
        placeholder="搜索规则名称..."
        style="width: 250px; margin-right: 16px"
        clearable
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select placeholder="规则类型" style="width: 150px; margin-right: 16px" clearable>
        <el-option label="路由规则" value="路由规则" />
        <el-option label="安全规则" value="安全规则" />
        <el-option label="转接规则" value="转接规则" />
        <el-option label="时间规则" value="时间规则" />
      </el-select>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增规则
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" :loading="loading" stripe>
      <el-table-column prop="ruleName" label="规则名称" width="180" />
      <el-table-column prop="ruleCode" label="规则编码" width="160">
        <template #default="{ row }">
          <code class="rule-code">{{ row.ruleCode }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="ruleType" label="规则类型" width="100">
        <template #default="{ row }">
          <el-tag :type="getRuleTypeColor(row.ruleType)" size="small">
            {{ row.ruleType }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="规则描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="priority" label="优先级" width="80">
        <template #default="{ row }">
          <span class="priority-badge">P{{ row.priority }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="triggerCount" label="触发次数" width="100" sortable>
        <template #default="{ row }">
          <span class="trigger-count">{{ row.triggerCount }}</span>
        </template>
      </el-table-column>
      <el-table-column label="有效期" width="180">
        <template #default="{ row }">
          <span class="date-range">{{ row.effectiveTime }} ~ {{ row.expireTime }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="enabled" label="状态" width="80">
        <template #default="{ row }">
          <el-switch
            :model-value="row.enabled"
            @change="handleToggleEnabled(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleTest(row)">测试</el-button>
          <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-area">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next"
        :total="30"
        :page-sizes="[10, 20, 50]"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="currentRule ? '编辑规则' : '新增规则'"
      width="720px"
      destroy-on-close
      @closed="resetRuleForm"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleFormRules"
        label-width="88px"
      >
        <div class="form-grid">
          <el-form-item label="规则名称" prop="ruleName">
            <el-input v-model="ruleForm.ruleName" />
          </el-form-item>
          <el-form-item label="规则编码" prop="ruleCode">
            <el-input v-model="ruleForm.ruleCode" />
          </el-form-item>
          <el-form-item label="规则类型" prop="ruleType">
            <el-select v-model="ruleForm.ruleType">
              <el-option label="路由规则" value="路由规则" />
              <el-option label="安全规则" value="安全规则" />
              <el-option label="转接规则" value="转接规则" />
              <el-option label="时间规则" value="时间规则" />
            </el-select>
          </el-form-item>
          <el-form-item label="优先级" prop="priority">
            <el-input-number v-model="ruleForm.priority" :min="1" :max="99" />
          </el-form-item>
          <el-form-item label="生效时间" prop="effectiveTime">
            <el-date-picker v-model="ruleForm.effectiveTime" type="date" value-format="YYYY-MM-DD" />
          </el-form-item>
          <el-form-item label="过期时间" prop="expireTime">
            <el-date-picker v-model="ruleForm.expireTime" type="date" value-format="YYYY-MM-DD" />
          </el-form-item>
        </div>

        <el-form-item label="规则描述" prop="description">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-form-item label="规则条件" prop="conditions">
          <div class="json-editor">
            <div class="json-editor__toolbar">
              <el-button size="small" @click="fillExample">示例</el-button>
              <el-button size="small" @click="formatJsonField('conditions')">格式化</el-button>
            </div>
            <el-input
              v-model="ruleForm.conditions"
              type="textarea"
              :rows="8"
              spellcheck="false"
              placeholder='{"field":"confidence","op":"lt","value":0.7}'
            />
          </div>
        </el-form-item>

        <el-form-item label="规则动作" prop="actions">
          <div class="json-editor">
            <div class="json-editor__toolbar">
              <el-button size="small" @click="formatJsonField('actions')">格式化</el-button>
            </div>
            <el-input
              v-model="ruleForm.actions"
              type="textarea"
              :rows="5"
              spellcheck="false"
              placeholder='{"reason":"需要澄清","action":"TRANSFER_OR_CLARIFY"}'
            />
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="testDialogVisible"
      title="测试规则"
      width="760px"
      destroy-on-close
    >
      <el-form label-width="96px">
        <el-form-item label="用户输入">
          <el-input v-model="testForm.userMessage" type="textarea" :rows="2" />
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="场景">
            <el-input v-model="testForm.sceneType" />
          </el-form-item>
          <el-form-item label="意图编码">
            <el-input v-model="testForm.intentCode" />
          </el-form-item>
          <el-form-item label="意图名称">
            <el-input v-model="testForm.intentName" />
          </el-form-item>
          <el-form-item label="置信度">
            <el-input-number v-model="testForm.confidence" :min="0" :max="1" :step="0.01" />
          </el-form-item>
          <el-form-item label="情绪">
            <el-input v-model="testForm.emotion" />
          </el-form-item>
          <el-form-item label="知识数量">
            <el-input-number v-model="testForm.knowledgeCount" :min="0" :max="99" />
          </el-form-item>
        </div>
        <el-form-item label="可靠知识">
          <el-switch v-model="testForm.hasReliableKnowledge" />
        </el-form-item>
        <el-form-item label="上下文">
          <el-input v-model="testForm.contextText" type="textarea" :rows="4" spellcheck="false" />
        </el-form-item>
      </el-form>

      <div class="test-result">
        <div class="test-result__header">
          <el-tag v-if="testForm.result" :type="testForm.result.hit ? 'success' : 'info'">
            {{ testForm.result.hit ? '已命中' : '未命中' }}
          </el-tag>
          <span v-else class="test-result__empty">尚未运行</span>
        </div>
        <div v-if="testForm.result?.ruleHit" class="test-result__content">
          <div><strong>来源：</strong>{{ testForm.result.ruleHit.conditionSource }}</div>
          <div><strong>动作：</strong>{{ testForm.result.ruleHit.action }}</div>
          <div><strong>原因：</strong>{{ testForm.result.ruleHit.reason }}</div>
        </div>
      </div>

      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleRunTest">运行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.rule-config {
  height: 100%;
}

.action-bar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 8px;
}

.rule-code {
  background-color: #f5f5f5;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.priority-badge {
  display: inline-block;
  width: 24px;
  height: 24px;
  line-height: 24px;
  text-align: center;
  background-color: #1890ff;
  color: #fff;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.trigger-count {
  color: #52c41a;
  font-weight: 600;
}

.date-range {
  font-size: 12px;
  color: #666;
}

.pagination-area {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 16px;
}

.json-editor {
  width: 100%;

  &__toolbar {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-bottom: 8px;
  }
}

.test-result {
  padding: 12px;
  border-radius: 6px;
  background-color: #fafafa;

  &__header {
    min-height: 24px;
    margin-bottom: 8px;
  }

  &__empty {
    color: #999;
    font-size: 12px;
  }

  &__content {
    display: grid;
    gap: 6px;
    color: #555;
    font-size: 13px;
    line-height: 1.5;
    word-break: break-word;
  }
}
</style>
