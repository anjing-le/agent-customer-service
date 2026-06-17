# RAG Retrieval Guide

本文档说明当前教学版 RAG 检索：它是从关键词匹配走向生产级向量检索的过渡层，用于课堂解释“为什么这条知识会成为 evidence”。

## 定位

当前实现不是外部向量数据库，也不是生产级 rerank 服务。

它的目标是：

- 让 seed store 和 PostgreSQL store 共用同一套检索排序逻辑。
- 给每条 evidence 返回 `retrievalScore` 和 `retrievalReason`。
- 保留无证据兜底边界，避免弱相关内容被误当成证据。
- 为后续接入向量数据库、embedding、rerank 服务预留清晰替换点。

## Runtime Path

```text
SendMessage
  -> SearchKnowledge
  -> rankKnowledge
  -> Rule / Guardrail
  -> evidence answer OR safe fallback
```

实现位置：

- `internal/platform/store/retrieval.go`
- `Store.SearchKnowledge`
- `PostgresStore.SearchKnowledge`
- `KnowledgeArticle.retrievalScore`
- `KnowledgeArticle.retrievalReason`

## Scoring Signals

教学版 rerank 使用可解释的轻量信号：

| 信号 | 作用 |
|---|---|
| title | 用户问题包含知识标题时强加权 |
| category | 用户问题命中知识分类时中等加权 |
| tag | 用户问题命中标签时中等加权 |
| semantic-overlap | 基于中英文 token / 中文 bigram 的轻量重合度 |
| trustLevel | `HIGH` / `MEDIUM` 知识会有小幅加权 |

命中后会返回：

```json
{
  "retrievalScore": 173.5,
  "retrievalReason": "title, tag:退货, tag:运费, semantic-overlap"
}
```

## Safety Boundary

当前策略会过滤低分候选，避免“只重合一两个泛词”的知识被当成可靠 evidence。

因此：

- 命中可信知识时，Agent 可以回答并展示 evidence。
- 没有足够证据时，Agent 仍然走 `NO_EVIDENCE` 兜底并创建 `KnowledgeGap`。
- 转人工规则优先级仍然高于自由回答。

## Teaching Script

课堂里可以这样讲：

1. 先问退货或发票问题，观察控制台 RAG 证据里的 score 和 reason。
2. 再问新品保价等知识库没有的问题，观察系统不会因为弱重合而强行回答。
3. 说明当前是教学版 lightweight rerank，生产环境会替换成 embedding + vector database + rerank service。

## V3 Replacement

生产化时替换点应保持在 `SearchKnowledge` / `rankKnowledge` 边界：

- 用 embedding 生成 query vector。
- 用向量数据库召回候选知识。
- 用 rerank 服务重排候选。
- 保留 `retrievalScore` / `retrievalReason`，但 reason 可以变成 `vector`, `rerank`, `policy-filter` 等。
- 继续保留低置信度兜底和 `KnowledgeGap`。
