# Domain Model

本文档描述 `agent-customer-service` 当前 Go runtime 的可靠客服 Agent 领域模型。

## Core Concepts

| 模型 | 说明 | 当前来源 |
|---|---|---|
| `Conversation` | 一条客服会话，记录客户、渠道、意图、风险和最后一句话 | `internal/platform/store` |
| `Message` | 用户或助手的一条消息，包含回复引擎、证据、兜底原因和 trace | `SendMessage` runtime |
| `AgentTrace` | 单轮回复的策略、证据数量、历史数量、模型尝试与回退观测 | `Message.trace` |
| `KnowledgeArticle` | 可引用的可信知识 | seed store 或 PostgreSQL |
| `KnowledgeGap` | 无可靠证据时沉淀的知识缺口 | `NO_EVIDENCE_FALLBACK` |
| `Rule` | 规则兜底或转人工策略 | seed store 或 PostgreSQL |
| `RuleTestResult` | 对单句用户输入的规则测试结果 | `/api/ops/rules/test` |
| `TransferTicket` | 转人工工单，包含 SLA、升级状态、创建和处理事件时间线 | `TRANSFER_THRESHOLD` |
| `TransferEvent` | 人工工单的创建、解决等留痕事件 | `TransferTicket.events` |
| `QualitySummary` | 质量评估摘要，统计证据回答、安全兜底、转人工和人工复核备注 | `/api/ops/dashboard` |
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
| 缺口处理 | 可以关闭缺口，也可以由缺口生成 `KnowledgeArticle` |
| 规则测试 | 不发送真实消息，也能验证转人工、无证据兜底和可回答边界 |

## Application Ports

当前 `store.Runtime` 是路由层依赖的应用端口：

| 方法 | 职责 |
|---|---|
| `ListConversations` / `CreateConversation` | 会话管理 |
| `SendMessage` | 一轮用户输入的核心处理 |
| `ListKnowledge` / `SearchKnowledge` | 知识列表和检索 |
| `ResolveKnowledgeGap` | 关闭知识缺口 |
| `CreateArticleFromGap` | 由缺口生成可信知识 |
| `TestRule` | 规则兜底测试 |
| `ResolveTransferTicket` | 人工工单处理 |
| `Dashboard` | 运营聚合视图 |

实现：

- `Store`: in-memory seed runtime，用于课堂演示和无数据库启动。
- `PostgresStore`: PostgreSQL runtime，用于真实持久化模式。

## Data Ownership

| 模块 | 主要数据 |
|---|---|
| `internal/customer` | conversation、message |
| `internal/knowledge` | knowledge article、knowledge gap |
| `internal/ops` | dashboard、rule test、transfer ticket |
| `internal/platform/store` | runtime interface、seed store、Postgres store |

## Current Limits

- 检索仍是轻量关键词匹配，V3 再引入向量检索和 rerank。
- 规则引擎是确定性轻量规则，尚未接完整表达式 DSL。
- 当前没有鉴权、多租户和限流。
- 模型客户端默认关闭，开启后只在有知识证据的路径参与生成；失败会自动回退到 `rag+rule`。
