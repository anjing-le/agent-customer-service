# Final Acceptance

本文档用于最终收口：确认 `agent-customer-service` 已经是一个可教学、可运行、可验证的可靠 Agent 客服业务样例。

## 完成定义

项目完成的标准不是功能越堆越多，而是能清楚讲明：

- 工程习惯来自 DVSkyFolding：Go、React/Vite、PostgreSQL、contracts、scripts、quality gate。
- 业务增量只关注可靠客服 Agent：多轮会话、RAG、规则兜底、防幻觉、会话历史和运营闭环。
- 无证据不自由生成，高风险可转人工，渠道请求不盲收。
- 每个关键动作都有 Dashboard、契约、测试或审计记录支撑。

## 必跑验收

```bash
git status --short --branch
pnpm check:delivery
pnpm verify
```

课堂 smoke 需要先启动 `go run ./cmd/platform-all`：

```bash
pnpm demo:classroom
```

也可以用一条命令构建 console、临时启动服务、跑 smoke 并自动停掉服务：

```bash
pnpm demo:classroom:local
```

可选 PostgreSQL 验收：

```bash
pnpm db:up
pnpm db:migrate
ANJING_INTEGRATION_DATABASE_URL='postgres://anjing:anjing@localhost:54330/agent_customer_service?sslmode=disable' go test ./internal/platform/store -run TestPostgresStoreRuntime -count=1
```

## 必讲脚本

1. 脚手架继承
   说明项目没有重新发明工程结构，而是沿用 DVSkyFolding 的命令、目录、配置、日志、数据库、契约和质量门禁。

2. 客服主链路
   启动 `go run ./cmd/platform-all`，打开控制台，演示知识命中、无证据兜底、转人工和 `Message.trace`。

3. 知识和规则运营
   演示知识缺口、规则测试、canary 对比、审批发布、回滚、质检标注和复盘样本。

4. 渠道接入边界
   演示真实渠道 adapter 的字段归一、签名、来源、限流、timestamp window、external message id 和 replay protection。

5. 渠道验收和 Runbook
   演示验收审计、验收质量越线、通知策略、Runbook 生成、批量分派、阻塞/恢复、负责人负载和操作审计事件。

6. 日报交接
   生成 Markdown/CSV 日报，说明失败摘要、验收率、Runbook 进度、负责人负载、Runbook 审计事件和交接优先级。

## 验收清单

- [x] 根 README 能在一分钟内讲清项目定位、启动方式和演示路径。
- [x] `project_document/DEMO_FLOW.md` 能支撑一次完整课堂演示。
- [x] `project_document/TEACHING_DELIVERY_CHECKLIST.md` 能在交付前确认仓库、author、远端分支、质量门禁、课堂 smoke 和文档口径。
- [x] 控制台顶部课堂主线能直接串起脚手架基线、客服主链路、RAG/规则、渠道验收、Runbook 和日报交接。
- [x] 控制台可以完成客服主链路、渠道接入、Runbook 和日报演示。
- [x] `contracts/api-contract.json` 覆盖新增接口和领域对象。
- [x] PostgreSQL migrations 能表达当前持久化模型。
- [x] `pnpm verify` 通过。
- [x] 最新 commit author 是 `安静 <245548353+anjing-le@users.noreply.github.com>`。
- [x] `main` 和 `master` 已推送到 GitHub。

## 明确不做

这些属于后续生产化扩展，不影响当前教学完成度：

- 向量数据库和 rerank。
- 完整鉴权、多租户、权限和成本治理。
- 真实平台的全部官方回调字段覆盖。
- 完整工单系统和客服排班系统。
- 大规模监控告警平台集成。

## 当前判断

当前项目已经达到可靠 Agent 客服教学样例的完成标准，完成度约 96%。后续若继续演进，应优先进入生产化扩展，而不是继续堆叠课堂概念。
