# Project Constraints

本文档定义 `agent-customer-service` 的长期工程约束。项目必须保持安静 DVSkyFolding 技术基线，目标不是套模板，而是把“可靠 Agent 客服”做成清晰、可验证、可演进的工程系统。

## 总体边界

- 项目定位是智能客服：多轮会话、RAG 检索增强、规则兜底、防幻觉、会话管理、历史记录、知识运营和人工接管。
- 底层技术栈、目录习惯、契约、脚本和质量门禁继承 DVSkyFolding；业务项目只扩展客服 Agent 设计。
- V1 保持轻量单体，但用多个 Go command 表达运行边界。
- 不默认引入网关、注册中心、MQ、对象存储、Redis 或复杂链路追踪平台；重型能力只保留接口、配置样例和扩展点。
- 可信知识、知识缺口、规则测试和人工队列必须能被 API 与控制台验证。
- 旧 `backend/`、`frontend/` 仅作迁移参考，新功能不得继续写入旧运行面。

## 后端约束

- 后端使用 Go，优先 `net/http` / `ServeMux`。
- 数据访问使用 PostgreSQL + `pgx/v5` + SQL，不引入 ORM。
- 配置使用环境变量，不引入配置中心。
- 日志使用 `log/slog` JSON。
- API 返回 `httpjson.Response` envelope：`success/data/error`。
- Store 接口必须显式返回 `error`，路由层统一转 JSON error。
- 会话、知识、规则、人工队列要保持清晰模块边界，不把所有逻辑塞进单个 handler。
- LLM API Key、数据库密码等敏感配置只允许出现在本地 ignored 文件或环境变量中。

## 前端约束

- 前端使用 React + TypeScript + Vite，入口位于 `apps/console`。
- 页面直接服务客服运营流程：会话、证据、缺口、规则、人工队列。
- UI 保持工具型信息密度，避免营销页、卡片套娃和过重装饰。
- API 类型就近定义，后续接 OpenAPI typed client 后再统一生成。
- 生产环境 API 不指向 mock。

## 数据约束

- PostgreSQL migration 放在 `infra/postgres/migrations`。
- 本地数据库脚本放在 `infra/local` 和根 `package.json`。
- Seed runtime 可以用于演示，但 PostgreSQL runtime 必须保持同等业务能力。

## 质量门禁

每批 Go/React 改动至少运行：

```bash
go test ./...
pnpm build:console
```

数据库相关改动在 Docker 可用时运行：

```bash
pnpm db:up
pnpm db:migrate
pnpm test:postgres
```
