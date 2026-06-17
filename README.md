# agent-customer-service

可靠智能客服教学项目，基于安静 DVSkyFolding 脚手架重构。

它不是从零发明一套工程习惯，而是在脚手架的 Go、React/Vite、PostgreSQL、contracts、scripts 和 quality gate 之上，生长出一个“可靠 Agent 客服”业务样例。

## 业务边界

- 多轮会话、历史记录和客服坐席视图
- RAG 知识检索、无证据兜底和知识缺口
- 规则兜底、转人工、人工质检和复盘样本
- 渠道 inbound 适配、签名、限流、replay 和验收审计
- 通知、日报、Runbook、负责人负载和操作审计

## 技术栈

- Frontend: React + TypeScript + Vite
- Backend: Go + `net/http` / `ServeMux`
- Database: PostgreSQL + `pgx/v5` + SQL migrations
- Config: env first
- Logging: `log/slog` JSON
- Delivery: one image, multiple Go commands

## 目录

```text
apps/console        React 控制台
cmd/platform-all    本地一体化启动
cmd/customer-service-api
cmd/ops-api
cmd/console-web
cmd/migrate-db
internal/platform   配置、HTTP JSON、DB、seed/PostgreSQL store
internal/customer   会话与 Agent Runtime API
internal/knowledge  知识检索 API
internal/ops        运营看板、日报、Runbook API
infra/postgres      PostgreSQL migrations
contracts           API、服务边界和渠道协议契约
```

## 快速开始

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

默认地址：`http://localhost:10002`

PostgreSQL 模式：

```bash
pnpm db:up
pnpm db:migrate
export ANJING_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable'
go run ./cmd/platform-all
```

模型客户端默认关闭。开启后也只在命中可信知识证据时参与生成；无证据、高风险或模型失败仍走规则兜底。

## 演示路径

推荐按这条链路讲：

1. 客服对话：知识命中、无证据兜底、转人工。
2. 运营视图：知识缺口、质检任务、规则测试、复盘样本。
3. 渠道接入：adapter 字段归一、签名、来源、限流、replay。
4. 渠道验收：成功/失败审计、验收质量事件、通知策略。
5. Runbook 闭环：生成处置步骤、批量分派、阻塞/恢复、负责人负载、操作审计。
6. 日报交接：生成 Markdown/CSV 日报，查看历史和补偿事件。

详细流程见 [Demo Flow](./project_document/DEMO_FLOW.md)。

## 校验

```bash
pnpm verify
```

常用拆分命令：

```bash
go test ./...
pnpm build:console
./scripts/check-agent-regression.sh
./scripts/check-agent-quality.sh
```

PostgreSQL 集成测试：

```bash
pnpm db:up
ANJING_INTEGRATION_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable' go test ./internal/platform/store -run TestPostgresStoreRuntime -count=1
```

## 文档

- [项目状态](./project_document/STATUS.md)
- [脚手架继承](./project_document/SCAFFOLD_INHERITANCE.md)
- [领域模型](./project_document/DOMAIN_MODEL.md)
- [服务边界](./project_document/SERVICE_BOUNDARY_GUIDE.md)
- [API Contract](./project_document/API_CONTRACT_GUIDE.md)
- [本地启动](./project_document/LOCAL_STARTUP_GUIDE.md)
- [路线图](./project_document/ROADMAP.md)
