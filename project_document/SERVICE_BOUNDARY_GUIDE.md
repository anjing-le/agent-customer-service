# Service Boundary Guide

`contracts/service-boundaries.json` 是 `agent-customer-service` 的服务边界事实来源。

## Current Boundaries

| boundary | kind | basePath | 说明 |
|---|---|---|---|
| customer-service | runtime | `/api/customer-service` | 会话、消息、RAG 回复、无证据兜底、转人工触发 |
| knowledge | runtime | `/api/knowledge` | 知识列表、检索、知识缺口关闭、缺口生成知识 |
| ops | runtime | `/api/ops` | 运行看板、规则测试、人工 ticket 处理、质检标注 |

## Go Commands

| command | 说明 |
|---|---|
| `cmd/platform-all` | 本地一体化运行 API 和 console web |
| `cmd/customer-service-api` | 会话与知识运行面 |
| `cmd/ops-api` | 运营看板、规则测试、人工队列 |
| `cmd/console-web` | 静态控制台服务 |
| `cmd/migrate-db` | PostgreSQL migration |

## Runtime Truth

- 用户消息进入 `/api/customer-service/messages`。
- 控制台可以使用 `/api/customer-service/messages/stream` 获取 SSE 流式输出。
- 有可信知识时返回 RAG 证据回复；配置模型客户端后可生成 `llm+rag`。
- 无可信知识时返回安全兜底，并创建知识缺口。
- 命中投诉、催办、法律风险或人工诉求时创建人工 ticket。
- 人工 ticket 返回创建和解决事件，便于复盘接管链路。
- 人工 ticket 会返回 SLA 状态、等待分钟数和升级标记，控制台可筛选升级工单。
- 人工质检可对助手消息提交标注，按证据贴合、安全性和帮助性回写质量摘要。
- 每条助手消息通过 `trace` 暴露策略、证据、历史和模型回退观测。
- 知识缺口可以关闭，也可以生成知识后回归到可检索状态。
- 规则测试可以在不发送真实消息的情况下验证兜底边界。

## Update Rules

- 新增运行接口先更新 `contracts/service-boundaries.json` 和 `contracts/api-contract.json`。
- 同步 React 控制台调用点。
- 运行 `go test ./...` 和 `pnpm build:console`。
