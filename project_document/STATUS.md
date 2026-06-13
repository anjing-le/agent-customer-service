# Status

更新时间：2026-06-13

## 当前状态

项目已从旧 `infra-dev-scaffolding` Java/Vue 方向切换到安静 DVSkyFolding 技术基线，第一版新骨架已经落地。

| 领域 | 状态 | 说明 |
|---|---|---|
| 技术基线 | In migration | 新增 Go module、React/Vite console、PostgreSQL migration，后续逐步退场旧 `backend/` 和 `frontend/` |
| Go 后端 | Skeleton ready | 使用 `net/http` / `ServeMux`、`log/slog`、统一 JSON envelope、env 配置 |
| 服务命令 | Skeleton ready | `platform-all`、`customer-service-api`、`ops-api`、`console-web`、`migrate-db` |
| 客服 API | V1 seed runtime | 会话列表、创建会话、发送消息、RAG seed 检索、无证据兜底、转人工兜底 |
| 知识 API | V1 seed runtime | 知识列表和关键词检索，后续接 PostgreSQL repository |
| 运营 API | V1 seed runtime | 运营指标、会话队列、知识缺口、规则清单 |
| 错误边界 | Runtime connected | Store 接口显式返回 error，路由层统一转 `store_error` JSON envelope |
| 前端控制台 | Runtime workspace | React/Vite 控制台已按对话中心、知识中心、场景配置迁移旧 Vue 信息架构，并展示证据、缺口和规则护栏 |
| 数据底座 | Runtime connected | `infra/postgres/migrations` 定义核心表和 demo seed；配置 `ANJING_DATABASE_URL` 后 API 切到 PostgreSQL store |
| 本地数据库 | Ready | `infra/local/docker-compose.yml`、`pnpm db:*` 脚本和可选 Postgres 集成测试已接入 |
| 知识运营闭环 | Runtime connected | 缺口支持关闭，支持由缺口生成可信知识条目并回写为已处理 |
| 规则测试 | Runtime connected | 场景配置区可输入问题测试转人工、无证据兜底和可回答边界 |

## 迁移原则

- 脚手架能力来自 DVSkyFolding：Go、React/Vite、PostgreSQL、SQL、env、JSON log、单镜像多 command。
- 客服 Agent 只新增业务设计，不重新发明底层工程习惯。
- 旧 Java/Vue 代码暂作历史参考，避免一次性删除导致教学材料断裂；后续每迁完一个边界就删除对应旧模块。
- 每个可验证增量都 commit/push。

## 下一步

1. 为 Go API contract 补机器可读文档。
2. 继续补齐 React 控制台的人工队列操作。
3. 删除旧 Java/Vue 运行入口，更新教学文档为 DVSkyFolding 口径。
