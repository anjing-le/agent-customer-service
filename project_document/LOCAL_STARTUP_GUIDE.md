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

## Quality Checks

```bash
./scripts/quality-gate.sh
```

常用拆分命令：

```bash
./scripts/check-agent-regression.sh
go test ./...
pnpm build:console
```
