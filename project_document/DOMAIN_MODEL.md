# Domain Model

本文档描述 `agent-customer-service` 当前 Go runtime 的可靠客服 Agent 领域模型。

## Core Concepts

| 模型 | 说明 | 当前来源 |
|---|---|---|
| `Conversation` | 一条客服会话，记录客户、渠道、意图、风险和最后一句话 | `internal/platform/store` |
| `ChannelInboundMessage` | 外部渠道进入系统的标准化消息，包含渠道、外部会话 ID、客户和内容 | `/api/channels/inbound` |
| `ChannelInboundReceipt` | 渠道入站流水，记录 replay key、外部消息 ID、签名、时间戳和内容 hash，用于防重复提交和渠道对账 | `channel_inbound_events` |
| `ChannelAdapterRequest` | 真实渠道回调的薄适配请求，例如 WeChat/WeCom/App/Marketplace 字段映射 | `/api/channels/*/inbound` 与 `contracts/channel-protocol-matrix.json` |
| `Message` | 用户或助手的一条消息，包含回复引擎、证据、兜底原因和 trace | `SendMessage` runtime |
| `AgentTrace` | 单轮回复的策略、证据数量、历史数量、模型尝试与回退观测 | `Message.trace` |
| `KnowledgeArticle` | 可引用的可信知识 | seed store 或 PostgreSQL |
| `KnowledgeGap` | 无可靠证据时沉淀的知识缺口 | `NO_EVIDENCE_FALLBACK` |
| `Rule` | 规则兜底或转人工策略，包含版本、阶段、命中次数和最后命中时间 | seed store 或 PostgreSQL |
| `RuleTestResult` | 对单句用户输入的当前规则测试结果 | `/api/ops/rules/test` |
| `RuleComparison` | 当前 active 规则与 canary 规则的处置结果对比 | `/api/ops/rules/compare` |
| `RuleApproval` | 规则发布审批记录，包含审批人、样本 ID 明细、自动样本数、风险等级和审批状态；控制台会把样本 ID 解析到 `TrainingSample` 详情 | `/api/ops/rules/approve` |
| `RuleReleaseEvent` | 规则发布/回滚事件，记录规则 code、版本、动作、操作者和备注 | `/api/ops/rules/*` |
| `RuleReleaseObservation` | 规则发布后的观测摘要，聚合命中、转人工、低分样本、人工均分和回滚建议 | `/api/ops/dashboard` |
| `TransferTicket` | 转人工工单，包含 SLA、升级状态、创建和处理事件时间线 | `TRANSFER_THRESHOLD` |
| `ChannelPolicy` | 渠道级客服策略，定义语气、风险加权、转人工 SLA 和升级说明 | seed store 或 PostgreSQL |
| `ChannelIntegration` | 渠道接入配置的非敏感视图，记录 active/next secret ref、allowed origins、rate limit、签名窗口和 replay 开关 | seed store 或 PostgreSQL |
| `ChannelAlert` | 渠道失败聚合，按渠道和错误码统计最近失败类型 | `Dashboard.channelAlerts` |
| `ChannelFailureTrend` | 渠道失败趋势桶，按渠道和小时统计最近失败量 | `Dashboard.channelFailureTrends` |
| `ChannelAlertPolicy` | 渠道失败通知策略，定义阈值、窗口、通知目标、target URL、secret ref、重试参数和当前触发状态 | `Dashboard.channelAlertPolicies` |
| `NotificationPolicyChange` | 高风险通知策略待审批变更，记录申请人、目标配置、变更前快照、字段级 diff、确认短语、过期时间、审批/拒绝/撤销/过期状态和处理结果 | `Dashboard.notificationPolicyChanges` |
| `NotificationPolicyEvent` | 通知策略变更审计事件，记录操作者、动作、改前/改后摘要、备注和时间 | `Dashboard.notificationPolicyEvents` |
| `ChannelNotification` | 渠道告警出站通知事件，记录目标 URL、secret ref、HMAC 签名、发送次数、退避重试、外部回执、投递审计摘要、死信原因和运营确认 | `Dashboard.channelNotifications` |
| `ChannelRunbook` | 渠道失败告警的运营处置步骤，由失败聚合、通知状态和告警策略派生，给出下一步、升级条件和检查项 | `Dashboard.channelRunbooks` |
| `ChannelOpsReport` | 渠道运营日报快照，保存 Markdown/CSV 正文、摘要指标、渠道范围和生成时间，用于审计、复盘和运营交接 | `/api/ops/channel-ops-reports/*` |
| `ReportSchedulerStatus` | 渠道运营日报后台调度状态，暴露开关、格式、间隔、保留数量、最近运行结果和下一次运行时间 | `/api/ops/channel-ops-report-scheduler` |
| `ChannelOpsReportEvent` | 渠道运营日报补偿事件，记录操作者、动作、结果、关联日报、prune 数量、备注和错误，支持按状态/操作者筛选和 CSV 导出 | `/api/ops/channel-ops-report-events` |
| `NotificationDeliveryAudit` | 通知投递审计记录，只保存 attempt、目标、secret ref、签名预览、payload hash 和脱敏请求/响应摘要，不保存完整 signed payload 或密钥值 | `ChannelNotification.deliveryAudit` |
| `TransferEvent` | 人工工单的创建、解决等留痕事件 | `TransferTicket.events` |
| `Annotation` | 对助手消息的人工质检标注，包含结论、备注、标签和三维评分 | `/api/ops/annotations/submit` |
| `AnnotationDimensions` | 人工质检评分维度：证据贴合、安全性、帮助性 | `Annotation.dimensions` |
| `ReviewTask` | 待复核任务，绑定助手消息、会话、渠道、优先级、领取人和完成时间 | `/api/ops/review-tasks/*` |
| `TrainingSample` | 由低分/待复核标注生成的复盘样本，包含 prompt、answer、证据、评分和备注 | `/api/ops/training-samples/export` |
| `QualitySummary` | 质量评估摘要，统计证据回答、安全兜底、转人工和人工标注均分 | `/api/ops/dashboard` |
| `Dashboard` | 会话、知识、规则、缺口和人工队列的运营聚合 | `/api/ops/dashboard` |

## Runtime Flow

```text
user message
  -> customer-service SendMessage
  -> keyword knowledge recall
  -> rule / guardrail decision
  -> evidence answer OR safe fallback
  -> create knowledge gap OR create transfer ticket when needed
  -> dashboard / console visibility
```

## Reliability Boundaries

| 场景 | 行为 |
|---|---|
| 命中可信知识 | 默认返回基于知识的 `rag+rule` 回复；配置模型客户端后可生成 `llm+rag` 回复 |
| 无可信知识 | 返回安全兜底，不自由生成；创建 `KnowledgeGap` |
| 投诉、催办、法律风险、人工诉求 | 返回转人工话术；创建 `TransferTicket` 和 `CREATED` 事件 |
| 渠道差异 | 按 `ChannelPolicy` 计算转人工 SLA，并在控制台按渠道筛选会话和工单 |
| 渠道接入 | adapter 先把真实渠道字段归一为标准 inbound，再读取 `ChannelIntegration` 做来源、限流、HMAC-SHA256 签名、时间窗、enabled、external message id 和 replay 校验 |
| 渠道观测 | Dashboard 聚合失败类型、小时趋势、通知策略、通知事件和处置 Runbook；通知事件可演示目标解析、secret ref 签名、高风险变更审批/拒绝/撤销/过期、审批人权限、二次确认短语、字段级 diff、通知目标回滚、demo/HTTP delivery client、退避重试、外部回执、死信和运营确认；日报调度状态单独暴露，补偿动作单独留痕，避免把后台任务状态混进实时 Dashboard |
| 缺口处理 | 可以关闭缺口，也可以由缺口生成 `KnowledgeArticle` |
| 规则测试 | 不发送真实消息，也能验证转人工、无证据兜底和可回答边界；正式测试只使用 `active` 规则并记录命中 |
| 规则灰度 | canary 规则不影响正式回复链路，先通过 `RuleComparison` 观察处置变化和人工队列压力 |
| 规则发布 | canary 需要先通过带样本明细的 `RuleApproval` 门禁，才能发布为 active；审批样本可关联低分复盘 `TrainingSample`，发布后进入 `RuleReleaseObservation` 观测并可回滚记录 `RuleReleaseEvent` |
| 质检任务 | 每条助手回复生成 `ReviewTask`，运营可领取/完成；提交 `Annotation` 会自动完成对应任务 |
| 人工质检 | 对助手消息提交 `Annotation`，把 groundedness、safety、helpfulness 汇总进质量评估 |
| 复盘导出 | 对低分、`FAIL` 或 `REVIEW` 标注生成 `TrainingSample`，用于运营复盘和后续训练数据整理 |

## Application Ports

当前 `store.Runtime` 是路由层依赖的应用端口：

| 方法 | 职责 |
|---|---|
| `ListConversations` / `CreateConversation` | 会话管理 |
| `SendMessage` | 一轮用户输入的核心处理 |
| `RecordChannelInbound` | 记录渠道入站 replay key，重复载荷不再进入客服链路 |
| `AcknowledgeChannelNotification` | 确认渠道通知事件 |
| `ReceiveChannelMessage` | 接收外部渠道消息并进入统一客服链路 |
| `ListKnowledge` / `SearchKnowledge` | 知识列表和检索 |
| `ResolveKnowledgeGap` | 关闭知识缺口 |
| `CreateArticleFromGap` | 由缺口生成可信知识 |
| `TestRule` | 规则兜底测试 |
| `CompareRuleVersions` | 对比 active 与 canary 规则处置结果 |
| `SubmitRuleApproval` | 提交规则发布审批门禁 |
| `PublishCanaryRule` / `RollbackRule` | 规则发布和回滚 |
| `ResolveTransferTicket` | 人工工单处理 |
| `AssignReviewTask` / `CompleteReviewTask` | 质检任务领取和完成 |
| `SubmitAnnotation` | 人工质检标注回写 |
| `ExportTrainingSamples` | 低分标注复盘样本导出 |
| `Dashboard` | 运营聚合视图 |

实现：

- `Store`: in-memory seed runtime，用于课堂演示和无数据库启动。
- `PostgresStore`: PostgreSQL runtime，用于真实持久化模式。

## Data Ownership

| 模块 | 主要数据 |
|---|---|
| `internal/customer` | conversation、message |
| `internal/channels` | channel adapters、inbound webhook、signature verification、replay protection、channel conversation mapping |
| `internal/knowledge` | knowledge article、knowledge gap |
| `internal/ops` | dashboard、rule test、transfer ticket、channel policy、channel integration、review task、annotation |
| `internal/platform/store` | runtime interface、seed store、Postgres store |

## Current Limits

- 检索仍是轻量关键词匹配，V3 再引入向量检索和 rerank。
- 规则引擎是确定性轻量规则，尚未接完整表达式 DSL。
- 当前没有鉴权、多租户和限流。
- 模型客户端默认关闭，开启后只在有知识证据的路径参与生成；失败会自动回退到 `rag+rule`。
- 渠道 adapter 已覆盖 WeChat/WeCom/App/Marketplace 的示例字段归一；密钥轮换支持 active/next secret ref 双密钥窗口，来源白名单支持 `Origin` / `X-Channel-Origin`。
