# Service Boundary Guide

`contracts/service-boundaries.json` 是 `agent-customer-service` 的服务边界事实来源。

## Current Boundaries

| boundary | kind | basePath | 说明 |
|---|---|---|---|
| customer-service | runtime | `/api/customer-service` | 会话、消息、RAG 回复、无证据兜底、转人工触发 |
| channels | runtime | `/api/channels` | 外部渠道 inbound 消息、签名校验、会话映射 |
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
- 外部渠道消息进入 `/api/channels/inbound`，读取 `ChannelIntegration` 后再做来源白名单、频率限制、HMAC-SHA256 签名、timestamp 时间窗和 replay 记录校验，并映射为内部会话。
- 真实渠道带 `externalMessageId` 时优先用 `channel + externalMessageId` 做幂等对账；没有消息 ID 时退回签名载荷 replay key。
- 微信、App、平台店铺 adapter 只做字段归一，统一复用标准 inbound 的验签、幂等和客服 Agent 链路。
- 控制台可以使用 `/api/customer-service/messages/stream` 获取 SSE 流式输出。
- 有可信知识时返回 RAG 证据回复；配置模型客户端后可生成 `llm+rag`。
- 无可信知识时返回安全兜底，并创建知识缺口。
- 命中投诉、催办、法律风险或人工诉求时创建人工 ticket。
- 人工 ticket 返回创建和解决事件，便于复盘接管链路。
- 人工 ticket 会返回 SLA 状态、等待分钟数和升级标记，控制台可筛选升级工单。
- 渠道策略会影响转人工 SLA，控制台可按 Web、WeChat、WeCom、App、Marketplace、Douyin、Xiaohongshu 筛选会话和工单。
- 渠道集成配置只暴露 active/next secret ref、allowed origins、签名窗口和 replay 开关，不把密钥值返回给控制台；运行时用 secret ref 解析 env secret，并允许双密钥轮换窗口。
- 渠道请求如果声明 `Origin` 或 `X-Channel-Origin`，必须命中 `allowedOrigins`，否则返回 `channel_origin_denied`。
- 渠道请求超过 `rateLimitPerMinute` 时返回 `channel_rate_limited`，不继续进入 Agent 回复链路。
- 人工质检可对助手消息提交标注，按证据贴合、安全性和帮助性回写质量摘要。
- 低分或待复核标注可导出为复盘样本，保留 prompt、answer、证据、评分和备注。
- 每条助手消息通过 `trace` 暴露策略、证据、历史和模型回退观测。
- 知识缺口可以关闭，也可以生成知识后回归到可检索状态。
- 规则测试可以在不发送真实消息的情况下验证兜底边界。

## Update Rules

- 新增运行接口先更新 `contracts/service-boundaries.json` 和 `contracts/api-contract.json`。
- 同步 React 控制台调用点。
- 运行 `go test ./...` 和 `pnpm build:console`。
