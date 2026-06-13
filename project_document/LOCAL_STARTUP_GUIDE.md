# Local Startup Guide

## Seed Runtime

不依赖数据库，适合课堂演示：

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

默认端口是 `10002`。打开 `http://localhost:10002` 可访问 React 控制台。

## PostgreSQL Runtime

```bash
pnpm db:up
pnpm db:migrate
export ANJING_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable'
go run ./cmd/platform-all
```

常用脚本：

```bash
pnpm db:logs
pnpm db:psql
pnpm db:down
```

## Optional Model Runtime

模型生成默认关闭。需要演示 `llm+rag` 路径时，配置 OpenAI-compatible Chat Completions 地址：

```bash
export ANJING_LLM_API_URL='https://example.com/v1/chat/completions'
export ANJING_LLM_API_KEY='your-api-key'
export ANJING_LLM_MODEL='gpt-4o-mini'
go run ./cmd/platform-all
```

模型只在命中知识证据时参与生成；无证据、高风险和模型失败都会回到规则兜底。

## Channel Webhook Runtime

`/api/channels/inbound` 使用 HMAC-SHA256 签名，校验 RFC3339 timestamp 是否在允许时间窗内，并记录 replay key 防止重复提交。默认保留 demo secret，真实渠道演示时用 env 覆盖：

```bash
export ANJING_CHANNEL_WEB_SECRET='your-web-secret'
export ANJING_CHANNEL_WECHAT_SECRET='your-wechat-secret'
export ANJING_CHANNEL_WECHAT_NEXT_SECRET='your-next-wechat-secret'
export ANJING_CHANNEL_APP_SECRET='your-app-secret'
export ANJING_CHANNEL_MARKETPLACE_SECRET='your-marketplace-secret'
export ANJING_CHANNEL_SIGNATURE_WINDOW_SECONDS=300
go run ./cmd/platform-all
```

真实渠道 adapter 入口：

- `/api/channels/wechat/inbound`
- `/api/channels/app/inbound`
- `/api/channels/marketplace/inbound`

adapter 只做字段归一，验签、时间窗、replay 和会话映射仍复用标准 inbound 链路。

## Smoke Checks

```bash
curl http://localhost:10002/healthz
curl http://localhost:10002/api/ops/dashboard
```

发送消息：

```bash
curl -X POST http://localhost:10002/api/customer-service/messages \
  -H "Content-Type: application/json" \
  -d '{"conversationId":"conv_demo_refund","content":"这个商品能不能开发票？"}'
```

规则测试：

```bash
curl -X POST http://localhost:10002/api/ops/rules/test \
  -H "Content-Type: application/json" \
  -d '{"content":"我已经投诉很多次了，现在必须转人工"}'
```

流式回复：

```bash
curl -N -X POST http://localhost:10002/api/customer-service/messages/stream \
  -H "Content-Type: application/json" \
  -d '{"conversationId":"conv_demo_refund","content":"这个商品能不能开发票？"}'
```

## Quality Checks

```bash
./scripts/quality-gate.sh
```

常用拆分命令：

```bash
./scripts/check-agent-regression.sh
./scripts/check-agent-quality.sh
go test ./...
pnpm build:console
```
