# API Contract Guide

机器可读版本见：

- `contracts/platform-contract.json`
- `contracts/service-boundaries.json`
- `contracts/api-contract.json`

## Response Envelope

Go API 当前响应格式：

```json
{
  "success": true,
  "data": {}
}
```

错误响应：

```json
{
  "success": false,
  "error": {
    "code": "store_error",
    "message": "error detail"
  }
}
```

## Current Runtime Boundaries

- `/api/customer-service/conversations`: 会话列表和创建。
- `GET /api/customer-service/messages?conversationId=...`: 读取指定会话历史消息。
- `POST /api/customer-service/messages`: 发送用户消息，触发 RAG、兜底、缺口和人工 ticket。
- `POST /api/customer-service/messages/stream`: 发送用户消息并以 SSE 返回 `meta`、`delta`、`done` 事件。
- `POST /api/channels/inbound`: 外部渠道消息入口，要求 `channel`、`externalConversationId`、RFC3339 `timestamp` 和 HMAC-SHA256 `signature`；签名密钥来自 env，timestamp 必须落在允许时间窗内。
- `/api/knowledge/articles`: 可信知识列表。
- `/api/knowledge/search`: 知识检索。
- `/api/knowledge/gaps/resolve`: 关闭知识缺口。
- `/api/knowledge/gaps/create-article`: 由知识缺口生成知识。
- `/api/ops/dashboard`: 运行看板。
- `/api/ops/rules/test`: 规则测试。
- `/api/ops/transfers/resolve`: 处理人工 ticket。
- `/api/ops/annotations/submit`: 对助手消息提交人工质检标注。
- `/api/ops/training-samples/export`: 导出低分或待复核标注生成的复盘样本。

## Endpoint Contract

`contracts/api-contract.json` 是当前接口的机器可读草案，覆盖：

- endpoint id、业务边界、method、path
- query/body request schema
- response status 和 data schema
- 客服 Agent 领域对象字段，例如 `Conversation`、`Message`、`ChannelInboundRequest`、`AgentTrace`、`KnowledgeArticle`、`KnowledgeGap`、`TransferTicket`、`TransferEvent`、`ChannelPolicy`、`Annotation`、`TrainingSample`、`QualitySummary`

新增或修改接口时，需要同时更新：

1. Go handler
2. React console 调用点
3. `contracts/service-boundaries.json`
4. `contracts/api-contract.json`

## Pagination

V1 数据量小，当前接口返回数组。后续列表增长后统一迁移到：

```json
{
  "records": [],
  "current": 1,
  "size": 20,
  "total": 100
}
```

## Verification

```bash
go test ./...
pnpm build:console
```
