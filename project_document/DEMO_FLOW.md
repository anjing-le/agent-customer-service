# Demo Flow

这条演示路径用于说明：底层工程习惯来自 DVSkyFolding，业务只关注可靠 Agent 客服设计。

## 1. 启动项目

```bash
pnpm install
pnpm build:console
go run ./cmd/platform-all
```

打开 `http://localhost:10002`。

## 2. 讲客服 Agent 主链路

在控制台发送客服问题，观察：

- 命中可信知识时，回复会携带 evidence。
- 没有证据时，系统不会自由发挥，会生成知识缺口。
- 投诉、催办、法律风险或人工诉求会进入转人工。
- `Message.trace` 会展示策略、证据数、历史数、耗时和兜底原因。

## 3. 讲知识、规则和质检闭环

按顺序演示：

- 知识缺口关闭或生成知识。
- 规则测试和 active/canary 对比。
- canary 发布审批、发布、回滚和观测。
- 人工质检标注，生成复盘样本。

## 4. 讲真实渠道接入

渠道 adapter 会先把 WeChat、WeCom、App、Marketplace、Douyin、Xiaohongshu 的平台字段归一为标准 inbound，再复用同一套校验：

- enabled
- allowed origins
- HMAC-SHA256 signature
- timestamp window
- rate limit
- external message id / replay protection

本地服务启动后可运行：

```bash
./scripts/demo-channel-inbound.sh
```

控制台也可以直接发送成功样例和失败样例，用于演示签名错误、来源不匹配、过期 timestamp、重复消息和限流。

## 5. 讲渠道验收和 Runbook

触发渠道失败后观察：

- `ChannelInboundAudit` 记录成功/拒绝验收，不保存完整 payload 或密钥。
- 验收质量越线会生成 `ChannelInboundAuditQualityEvent`。
- 渠道失败或验收异常会派生 `ChannelRunbook`。
- Runbook 检查项可以批量分派、完成、阻塞和恢复。
- Dashboard 会聚合负责人负载、逾期、阻塞和最近截止时间。
- 每次 ASSIGN、COMPLETE、BLOCK、RECOVER 都会生成 `ChannelRunbookCheckEvent`，进入 Dashboard、历史日报摘要和 CSV 导出。

## 6. 讲日报交接

在控制台生成 Markdown 或 CSV 渠道运营日报，观察：

- 渠道失败和验收率摘要。
- 验收质量 ACTIVE/WATCH/RECOVERED 汇总。
- Runbook 完成、阻塞、逾期和负责人负载。
- Runbook 审计事件计数和最近操作。
- 交接优先级建议和通知确认锚点。
- 历史日报、定时调度状态和补偿事件。

## 7. 收尾校验

```bash
pnpm verify
```

这一步会跑模板、contracts、渠道样例、Go 测试、Agent 回归、质量评估和前端 build。
