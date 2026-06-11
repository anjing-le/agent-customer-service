# Contracts

本目录存放 `agent-customer-service` 的机器可读工程契约。

- `platform-contract.json`: 平台级 API、响应、分页、请求上下文、时间、语言和错误码契约。
- `service-boundaries.json`: 当前单体内的服务/模块边界，以及未来拆分方向。

V1 先记录当前真实运行路径：`/api/chat`、`/api/knowledge`、`/api/scene`。后续如迁移到 `/api/customer-service/**`，需要先更新本目录，再同步后端 `ApiConstants`、前端 API 层和 OpenAPI 类型生成。
