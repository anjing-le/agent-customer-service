# API Contract Guide

机器可读版本见 `contracts/platform-contract.json` 和 `contracts/service-boundaries.json`。

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
- `/api/knowledge/articles`: 可信知识列表。
- `/api/knowledge/search`: 知识检索。
- `/api/knowledge/gaps/resolve`: 关闭知识缺口。
- `/api/knowledge/gaps/create-article`: 由知识缺口生成知识。
- `/api/ops/dashboard`: 运行看板。
- `/api/ops/rules/test`: 规则测试。
- `/api/ops/transfers/resolve`: 处理人工 ticket。

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
