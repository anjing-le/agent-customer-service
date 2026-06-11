# API Contract Guide

本文档记录 `agent-customer-service` 的 API 契约迁移方向。机器可读版本见 `contracts/platform-contract.json` 和 `contracts/service-boundaries.json`。

## Response Envelope

目标响应格式：

```json
{
  "code": "0",
  "message": "操作成功",
  "data": {},
  "timestamp": 1700000000000,
  "requestId": "request-id"
}
```

迁移规则：

- 新接口使用 `APIResponse<T>`。
- Chat/Knowledge/Scene 业务 Controller 已使用 `APIResponse<T>`。
- 非客服主链路的旧模板接口如仍返回 `R<T>`，属于兼容层，后续按模块替换。
- `code` 使用字符串，成功固定为 `"0"`。
- 前端页面只消费 `data`，错误、提示和链路 ID 由 HTTP 工具统一处理。
- 前端 HTTP 层兼容脚手架历史数字成功码 `200` 和客服契约字符串成功码 `"0"`。

## Pagination

目标分页结构：

```json
{
  "records": [],
  "current": 1,
  "size": 20,
  "total": 100
}
```

当前 `KnowledgeVO.PageVO` 和 `SceneVO.PageVO` 使用 `page` 字段，后续替换为共享 `PageResult<T>` 或字段兼容的 VO。

## V1 Runtime Boundaries

V1 先保留现有路径：

- `/api/chat/**`: 对话和会话运行接口。
- `/api/knowledge/**`: 知识管理和当前检索素材。
- `/api/scene/**`: 场景配置，当前未被对话运行链路消费。

目标路径见 `contracts/service-boundaries.json` 的 `targetBasePath`，后续迁移到 `/api/customer-service/**`。

前端运行路径统一维护在 `frontend/src/api/apiPaths.ts`，业务 API 文件不直接散落路径字符串。

## Verification

```bash
./scripts/check-contracts.sh
```
