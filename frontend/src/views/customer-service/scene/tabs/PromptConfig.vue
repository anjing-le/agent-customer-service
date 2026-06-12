<script setup lang="ts">
/**
 * 提示词模板管理 - 调用后端 SceneController 接口
 */
import {
  fetchListPrompts,
  fetchCreatePrompt,
  fetchUpdatePrompt,
  fetchDeletePrompt,
  fetchTestPrompt
} from '@/api/customer-service/scene'

interface PromptVariable {
  name: string
  description: string
  defaultValue: string
  required: boolean
}

interface Prompt {
  id: number
  promptName: string
  promptCode: string
  promptType: string
  content: string
  description: string
  variables: PromptVariable[]
  usageCount: number
  version: string
  status: string
  updatedAt: string
}

const tableData = ref<Prompt[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const testDialogVisible = ref(false)
const currentPrompt = ref<Prompt | null>(null)
const promptFormRef = ref()
const submitting = ref(false)
const promptForm = reactive({
  id: undefined as number | undefined,
  promptName: '',
  promptCode: '',
  promptType: 'SYSTEM',
  content: '',
  description: '',
  variables: [] as PromptVariable[],
  status: '启用'
})
const testForm = reactive({
  input: '请问这件衣服可以退货吗？',
  variablesText: '{}',
  renderedPrompt: '',
  aiResponse: ''
})

const runtimeVariables: PromptVariable[] = [
  { name: 'sceneType', description: '当前客服场景', defaultValue: '售后服务', required: false },
  { name: 'intentCode', description: '识别出的意图编码', defaultValue: 'RETURN_EXCHANGE', required: false },
  { name: 'intentName', description: '识别出的意图名称', defaultValue: '退换货咨询', required: false },
  { name: 'confidence', description: '意图置信度', defaultValue: '0.86', required: false },
  { name: 'emotion', description: '用户情绪', defaultValue: '中性', required: false },
  { name: 'userMessage', description: '用户原始输入', defaultValue: '请问这件衣服可以退货吗？', required: false },
  { name: 'knowledgeCount', description: '召回知识数量', defaultValue: '2', required: false },
  { name: 'hasReliableKnowledge', description: '是否有可靠知识', defaultValue: 'true', required: false }
]

const loadData = async () => {
  loading.value = true
  try {
    const res = await fetchListPrompts({})
    tableData.value = res.records || []
  } catch {
    // 后端不可用时保持空数据
  } finally {
    loading.value = false
  }
}

onMounted(() => { loadData() })

const handleAdd = () => {
  currentPrompt.value = null
  resetPromptForm()
  dialogVisible.value = true
}

const handleEdit = (row: Prompt) => {
  currentPrompt.value = row
  Object.assign(promptForm, {
    id: row.id,
    promptName: row.promptName || '',
    promptCode: row.promptCode || '',
    promptType: row.promptType || 'SYSTEM',
    content: row.content || '',
    description: row.description || '',
    variables: (row.variables || []).map(variable => ({ ...variable })),
    status: row.status || '启用'
  })
  dialogVisible.value = true
}

const handleTest = (row: Prompt) => {
  currentPrompt.value = row
  testForm.input = '请问这件衣服可以退货吗？'
  testForm.variablesText = JSON.stringify(
    Object.fromEntries((row.variables || []).map(variable => [variable.name, variable.defaultValue || ''])),
    null,
    2
  )
  testForm.renderedPrompt = ''
  testForm.aiResponse = ''
  testDialogVisible.value = true
}

const resetPromptForm = () => {
  Object.assign(promptForm, {
    id: undefined,
    promptName: '',
    promptCode: '',
    promptType: 'SYSTEM',
    content: '',
    description: '',
    variables: [],
    status: '启用'
  })
  promptFormRef.value?.clearValidate?.()
}

const variableNamePattern = /^[a-zA-Z][a-zA-Z0-9_]*$/

const validateVariables = (_rule: any, _value: PromptVariable[], callback: (error?: Error) => void) => {
  const names = new Set<string>()
  for (const variable of promptForm.variables) {
    if (!variable.name?.trim()) {
      callback(new Error('变量名不能为空'))
      return
    }
    if (!variableNamePattern.test(variable.name)) {
      callback(new Error('变量名只能使用字母、数字和下划线，且以字母开头'))
      return
    }
    if (names.has(variable.name)) {
      callback(new Error(`变量名重复：${variable.name}`))
      return
    }
    names.add(variable.name)
    if (variable.required && !promptForm.content.includes(`{{${variable.name}}}`)) {
      callback(new Error(`必填变量未出现在模板中：${variable.name}`))
      return
    }
  }
  callback()
}

const promptFormRules = {
  promptName: [{ required: true, message: '请输入提示词名称', trigger: 'blur' }],
  promptCode: [{ required: true, message: '请输入提示词编码', trigger: 'blur' }],
  promptType: [{ required: true, message: '请选择提示词类型', trigger: 'change' }],
  content: [{ required: true, message: '请输入提示词内容', trigger: 'blur' }],
  variables: [{ validator: validateVariables, trigger: 'change' }]
}

const addVariable = (variable?: PromptVariable) => {
  promptForm.variables.push(variable ? { ...variable } : {
    name: '',
    description: '',
    defaultValue: '',
    required: false
  })
}

const removeVariable = (index: number) => {
  promptForm.variables.splice(index, 1)
}

const insertVariable = (name: string) => {
  const token = `{{${name}}}`
  if (!promptForm.content.includes(token)) {
    promptForm.content = `${promptForm.content}${promptForm.content ? '\n' : ''}${token}`
  }
}

const applyRuntimeVariables = () => {
  const existing = new Set(promptForm.variables.map(variable => variable.name))
  runtimeVariables.forEach(variable => {
    if (!existing.has(variable.name)) addVariable(variable)
  })
}

const handleSubmit = async () => {
  await promptFormRef.value?.validate?.()
  submitting.value = true
  try {
    const payload = {
      ...promptForm,
      variables: promptForm.variables.map(variable => ({
        ...variable,
        name: variable.name.trim()
      }))
    }
    if (payload.id) {
      await fetchUpdatePrompt(payload)
    } else {
      await fetchCreatePrompt(payload)
    }
    ElMessage.success(payload.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleRunTest = async () => {
  if (!currentPrompt.value) return
  try {
    const variables = testForm.variablesText.trim() ? JSON.parse(testForm.variablesText) : {}
    const res = await fetchTestPrompt({
      promptId: currentPrompt.value.id,
      input: testForm.input,
      variables
    })
    testForm.renderedPrompt = res.renderedPrompt || ''
    testForm.aiResponse = res.aiResponse || ''
  } catch {
    ElMessage.warning('变量值必须是合法 JSON')
  }
}

const handleDelete = (row: Prompt) => {
  ElMessageBox.confirm('确定要删除该提示词模板吗？', '提示', {
    type: 'warning'
  }).then(async () => {
    await fetchDeletePrompt(row.id)
    ElMessage.success('删除成功')
    loadData()
  }).catch(() => {})
}

const getPromptTypeTag = (type: string) => {
  const typeMap: Record<string, { label: string; type: 'primary' | 'success' | 'warning' }> = {
    SYSTEM: { label: '系统', type: 'primary' },
    USER: { label: '用户', type: 'success' },
    ASSISTANT: { label: '助手', type: 'warning' }
  }
  return typeMap[type] || { label: type, type: 'primary' }
}
</script>

<template>
  <div class="prompt-config">
    <!-- 操作栏 -->
    <div class="action-bar">
      <el-input
        placeholder="搜索提示词名称..."
        style="width: 250px; margin-right: 16px"
        clearable
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增提示词
      </el-button>
    </div>

    <!-- 卡片列表 -->
    <div class="prompt-grid">
      <div
        v-for="item in tableData"
        :key="item.id"
        class="prompt-card"
      >
        <div class="prompt-card__header">
          <div class="prompt-card__title">
            <span class="prompt-card__name">{{ item.promptName }}</span>
            <el-tag :type="getPromptTypeTag(item.promptType).type" size="small">
              {{ getPromptTypeTag(item.promptType).label }}
            </el-tag>
          </div>
          <div class="prompt-card__version">{{ item.version }}</div>
        </div>

        <div class="prompt-card__code">
          <code>{{ item.promptCode }}</code>
        </div>

        <div class="prompt-card__desc">{{ item.description }}</div>
        <div class="prompt-card__variables">
          <el-tag
            v-for="variable in item.variables || []"
            :key="variable.name"
            size="small"
            type="info"
          >
            {{ variable.name }}
          </el-tag>
        </div>

        <div class="prompt-card__stats">
          <div class="stat-item">
            <span class="stat-value">{{ item.usageCount }}</span>
            <span class="stat-label">使用次数</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">{{ item.updatedAt?.split(' ')[0] || '-' }}</span>
            <span class="stat-label">更新日期</span>
          </div>
        </div>

        <div class="prompt-card__footer">
          <el-button type="primary" link size="small" @click="handleTest(item)">
            <el-icon><VideoPlay /></el-icon>
            测试
          </el-button>
          <el-button type="primary" link size="small" @click="handleEdit(item)">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button type="danger" link size="small" @click="handleDelete(item)">
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="currentPrompt ? '编辑提示词' : '新增提示词'"
      width="820px"
      destroy-on-close
      @closed="resetPromptForm"
    >
      <el-form
        ref="promptFormRef"
        :model="promptForm"
        :rules="promptFormRules"
        label-width="88px"
      >
        <div class="form-grid">
          <el-form-item label="名称" prop="promptName">
            <el-input v-model="promptForm.promptName" />
          </el-form-item>
          <el-form-item label="编码" prop="promptCode">
            <el-input v-model="promptForm.promptCode" />
          </el-form-item>
          <el-form-item label="类型" prop="promptType">
            <el-select v-model="promptForm.promptType">
              <el-option label="SYSTEM" value="SYSTEM" />
              <el-option label="USER" value="USER" />
              <el-option label="ASSISTANT" value="ASSISTANT" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态" prop="status">
            <el-select v-model="promptForm.status">
              <el-option label="启用" value="启用" />
              <el-option label="草稿" value="草稿" />
              <el-option label="停用" value="停用" />
            </el-select>
          </el-form-item>
        </div>

        <el-form-item label="描述" prop="description">
          <el-input v-model="promptForm.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-form-item label="内容" prop="content">
          <el-input
            v-model="promptForm.content"
            type="textarea"
            :rows="8"
            spellcheck="false"
          />
        </el-form-item>

        <el-form-item label="变量" prop="variables">
          <div class="variable-editor">
            <div class="variable-editor__toolbar">
              <el-button size="small" @click="applyRuntimeVariables">运行时变量</el-button>
              <el-button size="small" @click="addVariable()">新增变量</el-button>
            </div>
            <div
              v-for="(variable, index) in promptForm.variables"
              :key="`${variable.name}-${index}`"
              class="variable-row"
            >
              <el-input v-model="variable.name" placeholder="name" />
              <el-input v-model="variable.description" placeholder="description" />
              <el-input v-model="variable.defaultValue" placeholder="default" />
              <el-checkbox v-model="variable.required">必填</el-checkbox>
              <el-button size="small" @click="insertVariable(variable.name)">插入</el-button>
              <el-button size="small" type="danger" @click="removeVariable(index)">删除</el-button>
            </div>
            <div v-if="promptForm.variables.length === 0" class="empty-line">
              暂无变量
            </div>
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
      title="测试提示词"
      width="760px"
      destroy-on-close
    >
      <el-form label-width="88px">
        <el-form-item label="测试输入">
          <el-input v-model="testForm.input" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="变量值">
          <el-input v-model="testForm.variablesText" type="textarea" :rows="6" spellcheck="false" />
        </el-form-item>
      </el-form>
      <div class="test-result">
        <div class="test-result__block">
          <div class="test-result__title">渲染结果</div>
          <pre>{{ testForm.renderedPrompt || '-' }}</pre>
        </div>
        <div class="test-result__block">
          <div class="test-result__title">回复结果</div>
          <pre>{{ testForm.aiResponse || '-' }}</pre>
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
.prompt-config {
  height: 100%;
}

.action-bar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.prompt-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.prompt-card {
  background-color: #fafafa;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    border-color: #1890ff;
  }

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  &__title {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__name {
    font-size: 16px;
    font-weight: 600;
    color: #333;
  }

  &__version {
    font-size: 12px;
    color: #999;
    background-color: #f0f0f0;
    padding: 2px 8px;
    border-radius: 10px;
  }

  &__code {
    margin-bottom: 12px;

    code {
      background-color: #e6f7ff;
      color: #1890ff;
      padding: 4px 10px;
      border-radius: 4px;
      font-size: 12px;
    }
  }

  &__desc {
    font-size: 13px;
    color: #666;
    line-height: 1.6;
    margin-bottom: 16px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  &__variables {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    min-height: 24px;
    margin-bottom: 12px;
  }

  &__stats {
    display: flex;
    gap: 20px;
    padding: 12px;
    background-color: #fff;
    border-radius: 8px;
    margin-bottom: 16px;

    .stat-item {
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .stat-icon {
      font-size: 14px;
    }

    .stat-value {
      font-size: 14px;
      font-weight: 600;
      color: #333;
    }

    .stat-label {
      font-size: 11px;
      color: #999;
    }
  }

  &__footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding-top: 12px;
    border-top: 1px solid #f0f0f0;
  }
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 16px;
}

.variable-editor {
  width: 100%;

  &__toolbar {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-bottom: 8px;
  }
}

.variable-row {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr) 120px 64px 56px 56px;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.empty-line {
  color: #999;
  font-size: 12px;
}

.test-result {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;

  &__block {
    min-width: 0;
  }

  &__title {
    margin-bottom: 6px;
    color: #333;
    font-size: 13px;
    font-weight: 600;
  }

  pre {
    min-height: 120px;
    max-height: 220px;
    margin: 0;
    padding: 10px;
    overflow: auto;
    border-radius: 6px;
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
