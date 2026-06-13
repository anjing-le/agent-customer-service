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
| 前端控制台 | V1 shell | React/Vite 控制台已展示会话、Agent 工作区、知识缺口、兜底规则 |
| 数据底座 | Runtime connected | `infra/postgres/migrations` 定义核心表和 demo seed；配置 `ANJING_DATABASE_URL` 后 API 切到 PostgreSQL store |

## 迁移原则

- 脚手架能力来自 DVSkyFolding：Go、React/Vite、PostgreSQL、SQL、env、JSON log、单镜像多 command。
- 客服 Agent 只新增业务设计，不重新发明底层工程习惯。
- 旧 Java/Vue 代码暂作历史参考，避免一次性删除导致教学材料断裂；后续每迁完一个边界就删除对应旧模块。
- 每个可验证增量都 commit/push。

## 下一步

1. 为 Go API contract 补机器可读文档。
2. 将 PostgreSQL store 的错误返回显式化到 HTTP envelope。
3. 将旧 Vue 页面信息架构迁移到 React 控制台。
4. 删除旧 Java/Vue 运行入口，更新教学文档为 DVSkyFolding 口径。
