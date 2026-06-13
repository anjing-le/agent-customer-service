# Contracts

本目录存放 `agent-customer-service` 的机器可读工程契约。

- `platform-contract.json`: 平台级 API、响应、分页、请求上下文、时间、语言和错误码契约。
- `service-boundaries.json`: 当前 Go runtime 的服务/模块边界。
- `api-contract.json`: 当前 HTTP endpoint、request、response 和领域对象字段契约。

当前真实运行路径以 `service-boundaries.json` 为准：

- `/api/customer-service`
- `/api/knowledge`
- `/api/ops`

新增接口需要先更新本目录，再同步 Go handler 和 React 控制台调用点。
