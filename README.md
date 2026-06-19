# agent-customer-service

可靠 Agent 客服教学样例，基于安静 DVSkyFolding 脚手架重构。

底层工程习惯来自脚手架；业务只讲多轮会话、RAG evidence、无证据兜底、转人工、渠道验收和运营交接。

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

一条命令做课堂巡检：构建 console、临时启动服务、跑 smoke、自动停服。

```bash
pnpm install
pnpm demo:classroom:local
```

手动打开控制台：

```bash
pnpm build:console
go run ./cmd/platform-all
```

默认地址：`http://localhost:10002`

## 质量检查

```bash
pnpm verify
pnpm check:delivery
```

`pnpm verify` 跑模板、contracts、渠道样例、Go 测试、Agent 回归、质量评估和前端 build。

`pnpm check:delivery` 检查 Git author、`main/master` 远端同步和教学材料入口。

PostgreSQL、本地数据库和可选集成测试见 [本地启动](./project_document/LOCAL_STARTUP_GUIDE.md)。

## 关键目录

```text
apps/console        React 控制台
cmd/platform-all    本地一体化启动
internal/customer   会话与 Agent Runtime API
internal/knowledge  知识检索 API
internal/ops        运营看板、日报、Runbook API
internal/platform   配置、HTTP JSON、DB、store
infra/postgres      PostgreSQL migrations
contracts           API、服务边界和渠道协议契约
scripts             校验、回归和课堂 smoke
```

## 文档

- [课堂口播](./project_document/TEACHING_TALK_TRACK.md)
- [演示流程](./project_document/DEMO_FLOW.md)
- [最终验收](./project_document/FINAL_ACCEPTANCE.md)
- [交付体检](./project_document/TEACHING_DELIVERY_CHECKLIST.md)
- [脚手架继承](./project_document/SCAFFOLD_INHERITANCE.md)
- [领域模型](./project_document/DOMAIN_MODEL.md)
- [RAG 检索说明](./project_document/RAG_RETRIEVAL_GUIDE.md)
- [生产化扩展](./project_document/PRODUCTION_EXTENSION_GUIDE.md)
- [服务边界](./project_document/SERVICE_BOUNDARY_GUIDE.md)
- [项目状态](./project_document/STATUS.md)
