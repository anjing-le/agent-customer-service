# agent-customer-service

可靠智能客服教学项目，按安静 DVSkyFolding 脚手架口径重构：Go 后端、React + TypeScript + Vite 前端、PostgreSQL 数据底座、单镜像多 command 交付。

项目关注的业务边界是多轮会话、RAG 知识检索、规则兜底、防幻觉、会话管理、历史记录和运营质检。

## 技术栈

- Frontend: React + TypeScript + Vite
- Backend: Go + `net/http` / `ServeMux`
- Database: PostgreSQL + `pgx/v5` + SQL
- Logging: `log/slog` JSON
- Config: env first
- Delivery: one image, multiple Go commands

## 新工程结构

```text
apps/console        React 控制台
cmd/platform-all    本地一体化启动
cmd/customer-service-api
cmd/ops-api
cmd/console-web
cmd/migrate-db
internal/platform   配置、HTTP JSON、日志、DB、seed store
internal/customer   会话与 Agent Runtime API
internal/knowledge  知识检索 API
internal/ops        运营看板 API
infra/postgres      PostgreSQL migrations
```

早期运行面已退场，当前运行时只保留 DVSkyFolding 口径的 Go、React/Vite 和 PostgreSQL 结构。

## 快速开始

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

默认端口：后端和静态控制台 `10002`。

PostgreSQL 模式：

```bash
pnpm db:up
pnpm db:migrate
export ANJING_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable'
go run ./cmd/platform-all
```

可选模型生成模式：

```bash
export ANJING_LLM_API_URL='https://example.com/v1/chat/completions'
export ANJING_LLM_API_KEY='your-api-key'
export ANJING_LLM_MODEL='gpt-4o-mini'
go run ./cmd/platform-all
```

模型客户端默认关闭；开启后也只在命中知识证据时参与回复生成。无证据、高风险或模型不可用时仍走规则兜底。

流式回复：

```bash
curl -N -X POST http://localhost:10002/api/customer-service/messages/stream \
  -H "Content-Type: application/json" \
  -d '{"conversationId":"conv_demo_refund","content":"这个商品能不能开发票？"}'
```

## 校验

```bash
./scripts/quality-gate.sh
```

常用拆分命令：

```bash
./scripts/check-agent-regression.sh
go test ./...
pnpm build:console
```

PostgreSQL 集成测试：

```bash
pnpm db:up
ANJING_INTEGRATION_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable' go test ./internal/platform/store -run TestPostgresStoreRuntime -count=1
```

## 文档

- [项目状态](./project_document/STATUS.md)
- [领域模型](./project_document/DOMAIN_MODEL.md)
- [服务边界](./project_document/SERVICE_BOUNDARY_GUIDE.md)
- [路线图](./project_document/ROADMAP.md)
