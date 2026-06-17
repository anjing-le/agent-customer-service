# Status

更新时间：2026-06-14

## 当前状态

项目已从早期旧技术方向切换到安静 DVSkyFolding 技术基线，第一版新骨架已经落地。

| 领域 | 状态 | 说明 |
|---|---|---|
| 技术基线 | Runtime connected | Go module、React/Vite console、PostgreSQL migration 已按 DVSkyFolding 口径收敛 |
| Go 后端 | Runtime connected | 使用 `net/http` / `ServeMux`、`log/slog`、统一 JSON envelope、env 配置 |
| 服务命令 | Runtime connected | `platform-all`、`customer-service-api`、`ops-api`、`console-web`、`migrate-db` |
| 客服 API | Runtime connected | 会话列表、创建会话、发送消息、历史记录、RAG 检索、可选模型生成、无证据兜底、转人工兜底 |
| 知识 API | Runtime connected | 知识列表、关键词检索、知识缺口关闭、缺口生成知识，支持 seed 和 PostgreSQL |
| 运营 API | Runtime connected | 运营指标、会话队列、知识缺口、规则测试、人工队列 |
| 错误边界 | Runtime connected | Store 接口显式返回 error，路由层统一转 `store_error` JSON envelope |
| 前端控制台 | Runtime workspace | React/Vite 控制台已按客服坐席与运营工作台重新组织，并展示证据、缺口、规则护栏、渠道失败和通知投递治理 |
| 数据底座 | Runtime connected | `infra/postgres/migrations` 定义核心表和 demo seed；配置 `ANJING_DATABASE_URL` 后 API 切到 PostgreSQL store |
| 本地数据库 | Ready | `infra/local/docker-compose.yml`、`pnpm db:*` 脚本和可选 Postgres 集成测试已接入 |
| 知识运营闭环 | Runtime connected | 缺口支持关闭，支持由缺口生成可信知识条目并回写为已处理 |
| 规则测试 | Runtime connected | 场景配置区可输入问题测试转人工、无证据兜底和可回答边界 |
| 规则灰度 | Runtime connected | `Rule` 支持 active/canary 阶段、版本和命中计数，控制台可对比灰度规则是否改变处置结果；发布前需要带样本明细的审批门禁，审批记录可联动低分复盘样本详情，发布后生成观测指标和回滚建议，并可发布/回滚形成事件留痕 |
| 人工队列 | Runtime connected | 转人工回复自动生成 ticket，控制台可查看、筛选、处理工单，并展示 SLA、升级状态和事件时间线 |
| 质检任务 | Runtime connected | 助手回复自动生成待复核任务，控制台可领取/完成，提交人工标注后自动关闭对应任务 |
| 渠道策略 | Runtime connected | `ChannelPolicy` 定义 Web/WeChat/WeCom/App/Marketplace/Douyin/Xiaohongshu 的语气、风险加权和 SLA，转人工工单按渠道计算升级 |
| 渠道接入 | Runtime connected | `/api/channels/inbound` 读取 `ChannelIntegration` 后执行来源、限流、HMAC-SHA256 签名、timestamp 时间窗、enabled、external message id 和 replay 校验 |
| 渠道协议适配 | Runtime connected | `/api/channels/wechat|wecom|app|marketplace|douyin|xiaohongshu/inbound` 将真实渠道字段归一到标准 inbound 链路，控制台可发送成功/失败 contract examples，并展示 trace/evidence、平台验签 profile、协议差异和失败指标 |
| 渠道验收审计 | Runtime connected | `ChannelInboundAudit` 记录成功/拒绝入站的来源、验签结果、错误码、replay key、签名预览和 content hash，不保存完整 payload 或密钥；控制台支持按渠道、状态、错误码筛选、CSV 导出、验收质量条和越线恢复视图，日报汇总 accepted/rejected、验收率、高频错误码、越线次数、恢复渠道、仍需处理渠道、交接优先级建议和通知/Runbook 处理锚点，历史日报可直接确认关联通知，Runbook 检查项可逐条确认并保留 action ref，验收率过低或同一错误码高频出现时会按渠道策略阈值派生 Runbook，并沉淀 `ChannelInboundAuditQualityEvent` 复盘事件 |
| 渠道集成治理 | Runtime connected | `ChannelIntegration` 展示 active/next secret ref、allowed origins、rate limit、签名窗口、replay 开关和轮换提示，不返回密钥值 |
| 模型客户端 | Optional | 通过 `ANJING_LLM_*` 开启，只在有知识证据时生成；失败自动回退到规则 RAG |
| API 契约 | Runtime connected | `contracts/api-contract.json` 覆盖 endpoint、request、response 和领域对象字段，并纳入 contract 检查 |
| 回复观测 | Runtime connected | `Message.trace` 记录策略、证据数、历史数、模型尝试、耗时和回退原因；`ChannelAlert`、`ChannelFailureTrend`、`ChannelInboundAudit`、`ChannelInboundAuditQualityEvent`、`ChannelInboundAuditSummary`、`ChannelInboundAuditQualitySummary`、`ChannelOpsHandoffPriority`、`ChannelAlertPolicy`、`ChannelNotification`、`ChannelRunbook`、`ChannelRunbookCheck`、`ChannelOpsReport`、`ChannelOpsReportEvent`、`ReportSchedulerStatus`、`NotificationPolicyChange`、`NotificationPolicyEvent` 聚合渠道失败类型、趋势图、验收审计筛选导出、日报验收率、控制台验收质量条、越线恢复视图、日报越线摘要、运营交接优先级、历史日报通知确认动作、Runbook 检查项处置记录、批量分派、阻塞/恢复动作、逾期 dueAt 汇总、逾期筛选导出、负责人筛选、逾期交接优先级、日报处置摘要、负责人、截止时间、通知确认锚点、Runbook 复核锚点、按渠道配置的验收质量 Runbook、验收质量越线事件、通知策略、可配置目标 URL、secret ref、高风险变更审批/拒绝/撤销/过期、审批人权限、二次确认短语、字段级 diff、通知目标回滚、配置变更审计、demo/HTTP delivery client、签名出站、退避重试、外部回执、脱敏投递审计、死信、Runbook 处置、控制台 JSON 日报、服务端 Markdown/CSV 日报导出、手动/定时日报历史留存、保留策略、调度状态可视化、失败补偿留痕、事件筛选导出和确认闭环 |
| 回归验证 | Runtime connected | `testdata/agent_regression_cases.json` 固化可答、无证据、转人工、模型成功和模型失败回退场景 |
| 流式回复 | Runtime connected | `/api/customer-service/messages/stream` 以 SSE 返回 `meta`、`delta`、`done`，控制台已接入 |
| 质量评估 | Runtime connected | Dashboard 返回 `QualitySummary`，`testdata/quality_eval_cases.json` 固化质量评估用例并进入 quality gate |
| 人工质检 | Runtime connected | 支持对助手消息提交人工标注，记录 groundedness/safety/helpfulness 维度并汇总到 Dashboard 质量摘要 |
| 复盘导出 | Runtime connected | 低分或待复核标注可导出为 `TrainingSample`，控制台展示 prompt/answer/维度评分用于复盘 |

## 迁移原则

- 脚手架能力来自 DVSkyFolding：Go、React/Vite、PostgreSQL、SQL、env、JSON log、单镜像多 command。
- 客服 Agent 只新增业务设计，不重新发明底层工程习惯。
- 早期运行入口已移除，后续所有能力都在当前脚手架结构内演进。
- 每个可验证增量都 commit/push。

## 下一步

1. 补充 Runbook 分派审计事件和负责人负载统计。
2. 继续扩展更多平台渠道 profile。
