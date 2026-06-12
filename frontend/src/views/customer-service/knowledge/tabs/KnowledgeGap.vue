<script setup lang="ts">
/**
 * 知识缺口池 - 由对话运行时自动沉淀无答案问题。
 */
import { fetchListKnowledgeGaps, fetchResolveKnowledgeGap } from '@/api/customer-service/knowledge'

interface KnowledgeGap {
  id: number
  userQuestion: string
  intentName: string
  sceneType: string
  noAnswerReason: string
  noAnswerDetail: string
  status: string
  priority: string
  occurrenceCount: number
  lastSeenAt: string
  resolvedKnowledgeType?: string
  resolvedKnowledgeId?: number
  resolutionNote?: string
}

const searchKeyword = ref('')
const statusFilter = ref('OPEN')
const reasonFilter = ref('')
const tableData = ref<KnowledgeGap[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)
const currentGap = ref<KnowledgeGap | null>(null)
const formRef = ref()

const formData = reactive({
  question: '',
  answer: '',
  category: '售后服务',
  tags: '',
  resolutionNote: ''
})

const formRules = {
  answer: [{ required: true, message: '请输入补充答案', trigger: 'blur' }]
}

const priorityTagType = (priority?: string) => {
  if (priority === 'HIGH') return 'danger'
  if (priority === 'MEDIUM') return 'warning'
  return 'info'
}

const statusTagType = (status?: string) => status === 'RESOLVED' ? 'success' : 'warning'

const loadData = async () => {
  loading.value = true
  try {
    const res = await fetchListKnowledgeGaps({
      keyword: searchKeyword.value || undefined,
      status: statusFilter.value || undefined,
      noAnswerReason: reasonFilter.value || undefined,
      page: 1,
      size: 50
    })
    tableData.value = res.records || []
  } catch {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

const handleOpenResolve = (row: KnowledgeGap) => {
  currentGap.value = row
  formData.question = row.userQuestion
  formData.answer = ''
  formData.category = '售后服务'
  formData.tags = row.noAnswerReason ? row.noAnswerReason.toLowerCase().replaceAll('_', '-') : ''
  formData.resolutionNote = row.noAnswerDetail || ''
  dialogVisible.value = true
}

const handleCreateFaq = async () => {
  if (!currentGap.value) return
  try { await formRef.value?.validate() } catch { return }
  submitting.value = true
  try {
    await fetchResolveKnowledgeGap({
      id: currentGap.value.id,
      resolutionType: 'CREATE_FAQ',
      question: formData.question,
      answer: formData.answer,
      category: formData.category,
      tags: formData.tags ? formData.tags.split(',').map((t: string) => t.trim()).filter(Boolean) : [],
      resolutionNote: formData.resolutionNote
    })
    ElMessage.success('已补充为 FAQ')
    dialogVisible.value = false
    loadData()
  } catch {
    ElMessage.error('处理失败')
  } finally {
    submitting.value = false
  }
}

const handleManualClose = (row: KnowledgeGap) => {
  ElMessageBox.confirm('确定将该知识缺口标记为已处理吗？', '提示', { type: 'warning' }).then(async () => {
    await fetchResolveKnowledgeGap({
      id: row.id,
      resolutionType: 'MANUAL',
      resolutionNote: '人工确认无需新增知识'
    })
    ElMessage.success('已标记处理')
    loadData()
  }).catch(() => {})
}

onMounted(() => { loadData() })
</script>

<template>
  <div class="knowledge-gap">
    <div class="action-bar">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索用户问题或原因..."
        class="search-input"
        clearable
        @keyup.enter="loadData"
        @clear="loadData"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="statusFilter" class="filter-select" @change="loadData">
        <el-option label="待处理" value="OPEN" />
        <el-option label="已处理" value="RESOLVED" />
        <el-option label="全部" value="" />
      </el-select>
      <el-select v-model="reasonFilter" class="filter-select" clearable placeholder="原因" @change="loadData">
        <el-option label="无证据" value="NO_EVIDENCE" />
        <el-option label="低可信证据" value="LOW_TRUST_EVIDENCE" />
        <el-option label="未覆盖意图" value="UNSUPPORTED_INTENT" />
      </el-select>
      <el-button @click="loadData">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <el-table :data="tableData" :loading="loading" stripe>
      <el-table-column prop="userQuestion" label="用户问题" min-width="260" show-overflow-tooltip />
      <el-table-column prop="intentName" label="意图" width="120" show-overflow-tooltip />
      <el-table-column prop="noAnswerReason" label="无答案原因" width="160">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.noAnswerReason || 'NO_EVIDENCE' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="priority" label="优先级" width="90">
        <template #default="{ row }">
          <el-tag :type="priorityTagType(row.priority)" size="small">{{ row.priority || 'LOW' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="occurrenceCount" label="出现" width="80" sortable />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="noAnswerDetail" label="边界说明" min-width="260" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status !== 'RESOLVED'" type="primary" link size="small" @click="handleOpenResolve(row)">补 FAQ</el-button>
          <el-button v-if="row.status !== 'RESOLVED'" type="info" link size="small" @click="handleManualClose(row)">关闭</el-button>
          <span v-else class="resolved-text">{{ row.resolvedKnowledgeType || '已处理' }}</span>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="补充 FAQ" width="680px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="问题">
          <el-input v-model="formData.question" />
        </el-form-item>
        <el-form-item label="答案" prop="answer">
          <el-input v-model="formData.answer" type="textarea" :rows="5" placeholder="补充可被后续客服回复引用的标准答案" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分类">
              <el-select v-model="formData.category" style="width: 100%">
                <el-option label="售前咨询" value="售前咨询" />
                <el-option label="售后服务" value="售后服务" />
                <el-option label="物流配送" value="物流配送" />
                <el-option label="支付问题" value="支付问题" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="标签">
              <el-input v-model="formData.tags" placeholder="逗号分隔" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="处理说明">
          <el-input v-model="formData.resolutionNote" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCreateFaq">创建 FAQ 并关闭缺口</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.knowledge-gap {
  height: 100%;
}

.action-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.search-input {
  width: 320px;
}

.filter-select {
  width: 140px;
}

.resolved-text {
  color: #67c23a;
  font-size: 12px;
}
</style>
