# agent-customer-service

可靠智能客服教学项目，基于安静 DVSkyFolding 脚手架重构。

一句话：脚手架负责工程习惯，业务只讲“可靠 Agent 客服”设计。

## 课堂主线

1. 脚手架基线：Go、React/Vite、PostgreSQL、contracts、scripts、quality gate。
2. 客服主链路：多轮会话、历史记录、RAG evidence、无证据兜底、转人工。
3. 运营闭环：知识缺口、规则测试、人工质检、复盘样本。
4. 渠道接入：adapter 字段归一、签名、来源、限流、timestamp、replay。
5. 验收处置：入站审计、质量越线、通知策略、Runbook、负责人负载。
6. 日报交接：Markdown/CSV、历史日报、补偿事件、操作审计。

## 技术栈

- Frontend: React + TypeScript + Vite
- Backend: Go + `net/http` / `ServeMux`
- Database: PostgreSQL + `pgx/v5` + SQL migrations
- Config: env first
- Logging: `log/slog` JSON
- Delivery: one image, multiple Go commands

## 快速开始

Seed runtime 不依赖数据库，适合课堂演示：

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

默认地址：`http://localhost:10002`

课堂 smoke：

```bash
pnpm demo:classroom
```

完整质量门禁：

```bash
pnpm verify
```

PostgreSQL runtime：

```bash
pnpm db:up
pnpm db:migrate
export ANJING_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable'
go run ./cmd/platform-all
```

模型客户端默认关闭。开启后也只在命中可信知识证据时参与生成；无证据、高风险或模型失败仍走规则兜底。

## 关键目录

```text
apps/console        React 控制台
cmd/platform-all    本地一体化启动
internal/customer   会话与 Agent Runtime API
internal/knowledge  知识检索 API
internal/ops        运营看板、日报、Runbook API
internal/platform   配置、HTTP JSON、DB、seed/PostgreSQL store
infra/postgres      PostgreSQL migrations
contracts           API、服务边界和渠道协议契约
scripts             校验、回归和课堂 smoke
```

## 常用校验

- `pnpm verify`
- `go test ./...`
- `pnpm build:console`
- `./scripts/check-agent-regression.sh`
- `./scripts/check-agent-quality.sh`
- `pnpm demo:classroom`，需要先启动 `go run ./cmd/platform-all`

PostgreSQL 集成测试见 [本地启动](./project_document/LOCAL_STARTUP_GUIDE.md)。

## 文档

- [演示流程](./project_document/DEMO_FLOW.md)
- [最终验收](./project_document/FINAL_ACCEPTANCE.md)
- [脚手架继承](./project_document/SCAFFOLD_INHERITANCE.md)
- [领域模型](./project_document/DOMAIN_MODEL.md)
- [RAG 检索说明](./project_document/RAG_RETRIEVAL_GUIDE.md)
- [生产化扩展](./project_document/PRODUCTION_EXTENSION_GUIDE.md)
- [服务边界](./project_document/SERVICE_BOUNDARY_GUIDE.md)
- [项目状态](./project_document/STATUS.md)
