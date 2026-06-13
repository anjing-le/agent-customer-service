# Domain Model

本文档描述 `agent-customer-service` 当前 Go runtime 的可靠客服 Agent 领域模型。

## Core Concepts

| 模型 | 说明 | 当前来源 |
|---|---|---|
| `Conversation` | 一条客服会话，记录客户、渠道、意图、风险和最后一句话 | `internal/platform/store` |
| `ChannelInboundMessage` | 外部渠道进入系统的标准化消息，包含渠道、外部会话 ID、客户和内容 | `/api/channels/inbound` |
| `ChannelInboundReceipt` | 渠道入站流水，记录 replay key、签名、时间戳和内容 hash，用于防重复提交 | `channel_inbound_events` |
| `Message` | 用户或助手的一条消息，包含回复引擎、证据、兜底原因和 trace | `SendMessage` runtime |
| `AgentTrace` | 单轮回复的策略、证据数量、历史数量、模型尝试与回退观测 | `Message.trace` |
| `KnowledgeArticle` | 可引用的可信知识 | seed store 或 PostgreSQL |
| `KnowledgeGap` | 无可靠证据时沉淀的知识缺口 | `NO_EVIDENCE_FALLBACK` |
| `Rule` | 规则兜底或转人工策略 | seed store 或 PostgreSQL |
| `RuleTestResult` | 对单句用户输入的规则测试结果 | `/api/ops/rules/test` |
| `TransferTicket` | 转人工工单，包含 SLA、升级状态、创建和处理事件时间线 | `TRANSFER_THRESHOLD` |
| `ChannelPolicy` | 渠道级客服策略，定义语气、风险加权、转人工 SLA 和升级说明 | seed store 或 PostgreSQL |
| `TransferEvent` | 人工工单的创建、解决等留痕事件 | `TransferTicket.events` |
| `Annotation` | 对助手消息的人工质检标注，包含结论、备注、标签和三维评分 | `/api/ops/annotations/submit` |
| `AnnotationDimensions` | 人工质检评分维度：证据贴合、安全性、帮助性 | `Annotation.dimensions` |
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
| 渠道接入 | 对 inbound 请求做 HMAC-SHA256 签名、时间窗和 replay 校验，再映射到稳定内部会话 ID |
| 缺口处理 | 可以关闭缺口，也可以由缺口生成 `KnowledgeArticle` |
| 规则测试 | 不发送真实消息，也能验证转人工、无证据兜底和可回答边界 |
| 人工质检 | 对助手消息提交 `Annotation`，把 groundedness、safety、helpfulness 汇总进质量评估 |
| 复盘导出 | 对低分、`FAIL` 或 `REVIEW` 标注生成 `TrainingSample`，用于运营复盘和后续训练数据整理 |

## Application Ports

当前 `store.Runtime` 是路由层依赖的应用端口：

| 方法 | 职责 |
|---|---|
| `ListConversations` / `CreateConversation` | 会话管理 |
| `SendMessage` | 一轮用户输入的核心处理 |
| `RecordChannelInbound` | 记录渠道入站 replay key，重复载荷不再进入客服链路 |
| `ReceiveChannelMessage` | 接收外部渠道消息并进入统一客服链路 |
| `ListKnowledge` / `SearchKnowledge` | 知识列表和检索 |
| `ResolveKnowledgeGap` | 关闭知识缺口 |
| `CreateArticleFromGap` | 由缺口生成可信知识 |
| `TestRule` | 规则兜底测试 |
| `ResolveTransferTicket` | 人工工单处理 |
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
| `internal/channels` | inbound webhook、signature verification、replay protection、channel conversation mapping |
| `internal/knowledge` | knowledge article、knowledge gap |
| `internal/ops` | dashboard、rule test、transfer ticket、channel policy、annotation |
| `internal/platform/store` | runtime interface、seed store、Postgres store |

## Current Limits

- 检索仍是轻量关键词匹配，V3 再引入向量检索和 rerank。
- 规则引擎是确定性轻量规则，尚未接完整表达式 DSL。
- 当前没有鉴权、多租户和限流。
- 模型客户端默认关闭，开启后只在有知识证据的路径参与生成；失败会自动回退到 `rag+rule`。
- 渠道 replay 保护已覆盖重复签名载荷；下一步可以对接真实渠道 message id 和密钥轮换策略。
