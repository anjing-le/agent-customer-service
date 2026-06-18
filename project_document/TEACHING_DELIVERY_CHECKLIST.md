# Teaching Delivery Checklist

这份清单用于每次对外讲课、录屏或交付前做最终体检。目标不是证明功能越多越好，而是证明这个项目已经稳定表达：

- 底层工程习惯来自 DVSkyFolding。
- 业务增量聚焦可靠 Agent 客服。
- 运行、契约、测试、演示和文档能互相对上。

## 1. 仓库状态

确认本地没有遗漏改动，并且提交身份能计入 `anjing-le`：

```bash
git status --short --branch
git config --get user.name
git config --get user.email
git log -1 --pretty=fuller --date=iso
```

期望：

- worktree 干净。
- user.name 是 `安静`。
- user.email 是 `245548353+anjing-le@users.noreply.github.com`。
- 最新 commit 的 Author 和 Commit 都是这个身份。

确认本地、`main` 和 `master` 指向同一个交付版本：

```bash
git rev-parse HEAD origin/main origin/master
git ls-remote ssh://git@ssh.github.com:443/anjing-le/agent-customer-service.git refs/heads/main refs/heads/master
```

期望：

- `HEAD`、`origin/main`、`origin/master` hash 一致。
- 远端 `main` 和 `master` hash 一致。

## 2. 脚手架一致性

对外讲解前先确认项目仍然沿用 DVSkyFolding 技术栈：

| 层 | 应保持的习惯 | 检查位置 |
|---|---|---|
| Frontend | React + TypeScript + Vite | `apps/console` |
| Backend | Go + `net/http` / `ServeMux` | `cmd/*`、`internal/*` |
| Database | PostgreSQL + `pgx/v5` + SQL migrations | `infra/postgres/migrations` |
| Contract | 机器可读 API / service boundary | `contracts/` |
| Scripts | 一键质量门禁和课堂 smoke | `scripts/`、`package.json` |
| Config / Logs | env first、`log/slog` JSON | `internal/platform/config`、`internal/platform/service` |

如果新增能力需要引入其他基础设施，先写到 `PRODUCTION_EXTENSION_GUIDE.md`，不要直接塞进教学主链路。

## 3. 质量门禁

必跑：

```bash
pnpm verify
```

它应该覆盖：

- 模板和脚手架结构检查。
- API contract 和 service boundary 检查。
- 渠道协议 examples 检查。
- Go tests。
- Agent regression cases。
- Agent quality evaluation。
- React/Vite console build。

可选 PostgreSQL 验证：

```bash
pnpm db:up
pnpm db:migrate
pnpm test:postgres
```

## 4. 课堂演示

启动本地一体化服务：

```bash
go run ./cmd/platform-all
```

打开：

```text
http://localhost:10002
```

运行课堂 smoke：

```bash
pnpm demo:classroom
```

演示时至少讲清楚六件事：

1. 脚手架基线：Go、React/Vite、PostgreSQL、contracts、scripts、quality gate。
2. 客服主链路：多轮会话、RAG evidence、无证据兜底、转人工。
3. 知识运营：知识缺口、由缺口生成知识、规则测试、质检标注。
4. 渠道接入：adapter 字段归一、签名、来源、限流、timestamp、replay。
5. 运营恢复：验收审计、质量越线、通知策略、Runbook、负责人负载。
6. 日报交接：Markdown/CSV、历史日报、补偿事件、操作审计。

## 5. 文档口径

交付前按这个顺序读一遍，确认口径一致：

- `README.md`：一分钟说明项目定位、启动方式、演示路径。
- `project_document/SCAFFOLD_INHERITANCE.md`：说明它如何继承 DVSkyFolding。
- `project_document/DEMO_FLOW.md`：说明课堂怎么讲。
- `project_document/FINAL_ACCEPTANCE.md`：说明完成定义和验收命令。
- `project_document/STATUS.md`：说明当前完成度和剩余边界。
- `project_document/PRODUCTION_EXTENSION_GUIDE.md`：说明生产化扩展，不把 V3 复杂度压进教学样例。

## 6. 交付判断

满足以下条件即可认为当前版本可以交付：

- `pnpm verify` 通过。
- 课堂 smoke 通过。
- README、Demo Flow、Final Acceptance 和 Status 对项目定位的描述一致。
- 最新 commit author 是 `安静 <245548353+anjing-le@users.noreply.github.com>`。
- `main` 和 `master` 已推送到同一个最新 commit。
- 生产化工作已经被明确放到 V3 扩展章节，而不是混在当前课堂主线里。

如果其中任一项不满足，先修正这一项，再继续新增能力。
