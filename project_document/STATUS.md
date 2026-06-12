# Status

更新时间：2026-06-12

## 当前状态

| 领域 | 状态 | 说明 |
|---|---|---|
| 脚手架继承 | Ready | 技术栈、结构、契约、脚本、质量门禁沿用 `infra-dev-scaffolding`，业务只扩展客服 Agent |
| 对话中心 | Runtime observable | 会话/消息持久化仍在 `ChatService`，分析、检索、护栏、回复生成已由 `AgentRuntime` 编排；已补会话运行概览和单轮回复审计 |
| 知识中心 | In progress | Product/Activity/FAQ 参与当前对话检索；Industry/Solution 为预留 |
| 场景配置 | Runtime connected | 启用 Intent 参与关键词识别；启用 SYSTEM Prompt 由 PromptRuntime 渲染后注入 LLM 上下文；启用 Rule 由 RuleEngine 执行 |
| 工程契约 | Ready for V1 | 已接入 contracts、scripts、project_document；前端 API 路径已集中到 `ApiPaths` |
| 响应 envelope | Ready for V1 | Chat/Knowledge/Scene Controller 已迁移到 `APIResponse<T>`；前端 HTTP 层兼容字符串成功码 `"0"` |
| Agent 领域模型 | In runtime | `ConversationTurn`、`KnowledgeRecall`、`GuardrailDecision`、`AgentReply` 已接入对话主链路 |
| RuleEngine | Lightweight runtime | 支持内置规则码和轻量 JSON 条件表达式；命中后返回来源、原因和动作，并累加触发次数 |
| Rule 配置 | Runtime editable | 前端支持规则新增/编辑、条件/动作 JSON 格式化、表达式测试和校验；后端 create/update 拦截非法 JSON |
| PromptRuntime | Lightweight runtime | 支持 SYSTEM Prompt 场景过滤、基础变量渲染、使用次数累加，并返回渲染结果 |
| Prompt 配置 | Runtime editable | 前端支持 Prompt 新增/编辑、变量 schema、运行时变量快捷填充和测试；后端校验变量名、重复项和必填占位符 |
| Scene 运行概览 | Runtime observable | 展示启用意图、启用 SYSTEM Prompt、启用规则、规则命中、Prompt 使用、Top 项和运行洞察 |
| Chat 运行概览 | Runtime observable | 展示总会话、活跃会话、今日会话、今日消息、Agent 回复、安全回复、兜底回复、最近审计、7 日趋势和最近快照 |
| Agent 审计 | Runtime observable | 每轮助手回复沉淀意图、回复引擎、兜底原因、召回数、规则命中数和 Prompt 渲染数 |
| 会话质检摘要 | Runtime observable | 基于单轮审计聚合会话可靠性评分、风险等级、兜底率、不安全率和主要兜底原因 |
| Runtime Snapshot | Runtime observable | 支持手动采样和可配置定时采样，沉淀 Chat 运行质量历史快照 |
| 可靠性可观测 | In progress | 对话响应已返回回复引擎、兜底原因、护栏标签、命中规则和渲染提示词；前端 Agent 工作区已展示可靠性 Tab |
| OpenAPI typed client | Pending | 后续接入 `/v3/api-docs` 和前端 operation 类型 |
| 测试与质量门禁 | Ready for V1 | `./scripts/quality-gate.sh` 已通过；后续补单元测试和运行探针 |

## 当前保护边界

- 保留用户已有 README 收敛、OneRouter 地址更新和教学素材。
- 不读取或改动 ignored 的 `backend/src/main/resources/application-local.yml`。
- V2 第一阶段保持 API/前端 VO 不变，只替换后端内部编排结构。
- 业务运行链路优先收口，模板遗留接口后续按模块逐步清理。
- Scene 已补运行洞察；Chat 已沉淀 Agent 单轮审计事实、会话质检摘要、7 日趋势和运行快照。

## 验证入口

```bash
./scripts/check-template.sh
./scripts/check-contracts.sh
```

完整构建验证：

```bash
./scripts/quality-gate.sh
```
