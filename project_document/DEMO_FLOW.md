# Demo Flow

这条路径用于课堂演示：底层工程习惯来自 DVSkyFolding，业务只关注可靠 Agent 客服设计。

## 课前准备

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

打开 `http://localhost:10002`，然后跑一遍课堂 smoke：

```bash
pnpm demo:classroom
```

## 1. 先讲脚手架

讲法：这个项目不是重新发明工程结构，而是在脚手架的 Go、React/Vite、PostgreSQL、contracts、scripts、quality gate 上长业务。

看哪里：README、控制台顶部“课堂主线”、`contracts/`、`scripts/`、`infra/postgres/migrations/`。

证明什么：学生以后换业务，只需要关注领域设计，底层习惯保持一致。

## 2. 再讲客服主链路

讲法：可靠客服 Agent 的第一原则是“有证据才回答，没证据就兜底”。

看哪里：控制台“对话中心”和“RAG 证据”。

演示：

- 问退货运费，观察 evidence 和 `Message.trace`。
- 观察 evidence 的 `retrievalScore` / `retrievalReason`，说明当前是可解释的教学版 rerank。
- 问知识库没有的问题，观察知识缺口。
- 问投诉/催办/人工诉求，观察转人工。

## 3. 接着讲运营闭环

讲法：Agent 不只是生成回复，还要把不确定、低质量和高风险样本交给人处理。

看哪里：知识缺口、规则测试、质检任务、人工标注、复盘样本。

证明什么：RAG、规则、人工质检是边界，不是装饰。

## 4. 然后讲渠道接入

本地服务启动后可运行：

```bash
./scripts/demo-channel-inbound.sh
```

讲法：真实渠道先归一字段，再复用同一套验签、来源、限流、timestamp 和 replay 校验。

看哪里：渠道接入、协议样例、协议差异、失败演示。

证明什么：不同平台不是写一堆散乱入口，而是 adapter + 标准 inbound。

## 5. 最后讲验收、Runbook 和日报

讲法：渠道请求不能只看成功响应，还要能审计、恢复和交接。

看哪里：验收审计、验收质量、通知策略、Runbook、负责人负载、日报留存。

证明什么：

- `ChannelInboundAudit` 记录成功/拒绝，不保存完整 payload 或密钥。
- `ChannelRunbookCheckEvent` 记录 ASSIGN、COMPLETE、BLOCK、RECOVER。
- Markdown/CSV 日报能把失败、验收率、负责人负载和交接优先级带走。

## 收尾校验

```bash
pnpm verify
```

这一步会跑模板、contracts、渠道样例、Go 测试、Agent 回归、质量评估和前端 build。课堂前再跑 `pnpm demo:classroom`，确认本地服务和演示链路都通。
