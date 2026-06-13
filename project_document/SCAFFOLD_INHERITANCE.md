# Scaffold Inheritance

`agent-customer-service` 是按安静 DVSkyFolding 技术基线重构出来的可靠 Agent 客服样例。

教学目标是让学习者把注意力放在业务设计上：多轮会话、RAG、规则兜底、防幻觉、知识运营闭环、人工接管和运行时可观测。底层工程习惯、契约、质量门禁和交付方式来自脚手架。

## Inherited From DVSkyFolding

| 能力 | 脚手架约定 | 本项目状态 |
|---|---|---|
| 前端 | React + TypeScript + Vite，统一大前端入口 | `apps/console` 已落地客服 Agent 控制台 |
| 后端 | Go + `net/http` / `ServeMux`，不引入 Web 框架 | `cmd/*` 和 `internal/*` 已落地 |
| 数据库 | PostgreSQL + `pgx/v5` + SQL，不引入 ORM | `infra/postgres/migrations` 和 Postgres store 已接入 |
| 日志 | `log/slog` JSON 结构化日志 | `internal/platform/service` 已统一处理 |
| 配置 | env first | `ANJING_ADDR`、`ANJING_DATABASE_URL`、`ANJING_CONSOLE_DIST`、`ANJING_LLM_*` |
| 交付 | 单镜像多 command；本地可 `platform-all` | `cmd/platform-all`、`customer-service-api`、`ops-api`、`console-web`、`migrate-db` |
| 文档 | 状态、路线图、边界、启动、约束分开维护 | `project_document` 已按新基线更新 |

## Business Added By This Project

- `customer`: 会话、消息、Agent 回复、知识证据和转人工触发。
- `knowledge`: 可信知识、知识缺口、缺口关闭、由缺口生成知识。
- `ops`: 运行看板、规则测试、人工队列处理。
- `platform/store`: seed runtime 与 PostgreSQL runtime 的统一接口。
- `apps/console`: 对话中心、知识中心、场景配置、人工队列的 React 控制台。

## Runtime Structure

当前运行时以 DVSkyFolding 目录为准：

```text
apps/console
cmd/*
internal/*
infra/postgres
infra/local
```

早期运行入口已退场，教学材料只围绕当前脚手架结构展开。

## Teaching Narrative

1. 先看脚手架：Go command、React console、PostgreSQL migration、env 配置和 JSON 日志。
2. 再看边界：客服 Agent 只新增会话、知识、规则、人工队列。
3. 再看链路：用户消息如何触发 RAG、兜底、缺口、人工 ticket。
4. 再看可靠性：没有证据不回答，转人工生成工单，缺口可以补知识。
5. 最后看扩展：可选模型客户端、向量检索、流式回复、多租户和更完整的质检体系。

## Verification

每批 Go/React 改动至少运行：

```bash
go test ./...
pnpm build:console
```

PostgreSQL 集成验证在本地数据库可用时运行：

```bash
pnpm db:up
pnpm db:migrate
pnpm test:postgres
```
