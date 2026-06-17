# API Contract Guide

机器可读版本见：

- `contracts/platform-contract.json`
- `contracts/service-boundaries.json`
- `contracts/api-contract.json`
- `contracts/channel-protocol-matrix.json`
- `contracts/examples/channel-protocols.json`

## Response Envelope

Go API 当前响应格式：

```json
{
  "success": true,
  "data": {}
}
```

错误响应：

```json
{
  "success": false,
  "error": {
    "code": "store_error",
    "message": "error detail"
  }
}
```

## Current Runtime Boundaries

- `/api/customer-service/conversations`: 会话列表和创建。
- `GET /api/customer-service/messages?conversationId=...`: 读取指定会话历史消息。
- `POST /api/customer-service/messages`: 发送用户消息，触发 RAG、兜底、缺口和人工 ticket。
- `POST /api/customer-service/messages/stream`: 发送用户消息并以 SSE 返回 `meta`、`delta`、`done` 事件。
- `POST /api/channels/inbound`: 外部渠道消息入口，要求 `channel`、`externalConversationId`、RFC3339 `timestamp` 和 HMAC-SHA256 `signature`；可选 `externalMessageId` 用于真实渠道消息对账。校验会读取 `ChannelIntegration` 的 enabled、secretRef、allowed origins、rate limit、签名时间窗和 replay 开关，重复 `externalMessageId` 或签名载荷返回 `409 duplicate_inbound`，超过频率限制返回 `429 channel_rate_limited`。
- `POST /api/channels/wechat/inbound`: 微信协议 adapter，将 `openId`、`msgId`、`text` 归一为标准 inbound 请求。
- `POST /api/channels/wecom/inbound`: 企业微信协议 adapter，将 `corpId + userId`、`msgId`、`text` 归一为标准 inbound 请求。
- `POST /api/channels/app/inbound`: App 协议 adapter，将 `deviceId`、`messageId`、`body` 归一为标准 inbound 请求。
- `POST /api/channels/marketplace/inbound`: 平台店铺协议 adapter，将 `buyerId`、`eventId`、`message` 归一为标准 inbound 请求。
- `/api/knowledge/articles`: 可信知识列表。
- `/api/knowledge/search`: 知识检索。
- `/api/knowledge/gaps/resolve`: 关闭知识缺口。
- `/api/knowledge/gaps/create-article`: 由知识缺口生成知识。
- `/api/ops/dashboard`: 运行看板，包含规则发布后的 `RuleReleaseObservation[]` 观测摘要和回滚建议。
- `/api/ops/rules/test`: 规则测试。
- `/api/ops/rules/compare`: 对比 active 与 canary 规则版本的处置差异。
- `/api/ops/rules/approve`: 提交 canary 发布审批门禁，记录样本 ID 明细、自动样本数和风险等级；控制台会用 `sampleIds` 关联 `/api/ops/training-samples/export` 的复盘样本详情。
- `/api/ops/rules/publish-canary`: 将 canary 规则发布为 active 并记录事件。
- `/api/ops/rules/rollback`: 回滚 active 规则并记录事件。
- `/api/ops/transfers/resolve`: 处理人工 ticket。
- `/api/ops/review-tasks/assign`: 领取助手回复质检任务。
- `/api/ops/review-tasks/complete`: 完成助手回复质检任务。
- `/api/ops/channel-notifications/dispatch`: 渠道告警通知出站，解析通知目标 URL 和 secret ref，并通过可替换 delivery client 生成 HMAC-SHA256 签名；默认 demo client 可稳定教学，配置 `ANJING_NOTIFICATION_DELIVERY_MODE=http` 后使用真实 HTTP client；失败会按退避时间重试，成功记录外部回执，超过最大次数进入死信。
- `/api/ops/channel-inbound-audits`: 查询渠道入站验收审计，支持渠道、状态和错误码筛选。
- `/api/ops/channel-inbound-audits/export`: 导出渠道入站验收审计 CSV。
- `/api/ops/channel-inbound-audit-quality-events`: 查询渠道验收质量越线事件，支持渠道、状态和错误码筛选。
- `/api/ops/channel-inbound-audit-quality-events/export`: 导出渠道验收质量越线事件 CSV。
- `/api/ops/annotations/submit`: 对助手消息提交人工质检标注。
- `/api/ops/training-samples/export`: 导出低分或待复核标注生成的复盘样本。
- `/api/ops/channel-ops-report/export`: 导出渠道运营日报，`format=markdown` 返回 Runbook 报告，`format=csv` 返回表格摘要。
- `/api/ops/channel-ops-report-scheduler`: 查看渠道运营日报后台调度配置、最近一次运行状态和下一次运行时间。
- `/api/ops/channel-ops-report-scheduler/compensate`: 人工补生成渠道运营日报，并记录补偿审计事件。
- `/api/ops/channel-ops-report-events`: 查询渠道运营日报补偿事件。
- `/api/ops/channel-ops-reports`: 查询最近生成的渠道运营日报历史。
- `/api/ops/channel-ops-reports/generate`: 基于当前 Dashboard 生成并留存 Markdown/CSV 渠道运营日报。
- `/api/ops/channel-ops-reports/export`: 按日报 ID 下载历史 Markdown/CSV 渠道运营日报。
- `/api/ops/channel-runbook-check-events`: 查询 Runbook 检查项操作审计事件，支持渠道、状态、动作、操作者、负责人和 action ref 筛选。
- `/api/ops/channel-runbook-check-events/export`: 导出 Runbook 检查项操作审计 CSV。

## Endpoint Contract

`contracts/api-contract.json` 是当前接口的机器可读草案，覆盖：

- endpoint id、业务边界、method、path
- query/body request schema
- response status 和 data schema
- 客服 Agent 领域对象字段，例如 `Conversation`、`Message`、`ChannelInboundRequest`、`WeChatInboundRequest`、`WeComInboundRequest`、`AppInboundRequest`、`MarketplaceInboundRequest`、`DouyinInboundRequest`、`XiaohongshuInboundRequest`、`AgentTrace`、`KnowledgeArticle`、`KnowledgeGap`、`RuleComparison`、`RuleApproval`、`RuleReleaseEvent`、`RuleReleaseObservation`、`TransferTicket`、`TransferEvent`、`ChannelPolicy`、`ChannelIntegration`、`ReviewTask`、`Annotation`、`TrainingSample`、`QualitySummary`；`ChannelIntegration` 只暴露 active/next secret ref、allowed origins 和 rate limit，不暴露密钥值。

新增或修改接口时，需要同时更新：

1. Go handler
2. React console 调用点
3. `contracts/service-boundaries.json`
4. `contracts/api-contract.json`

## Channel Protocol Examples

`contracts/examples/channel-protocols.json` 固化了标准 inbound、WeChat、WeCom、App、Marketplace、Douyin 和 Xiaohongshu adapter 的请求样例，并为每个真实平台补充 `platformSignatureProfiles`。每个请求样例都包含：

- endpoint 和请求头
- demo secret 与 secret ref
- 参与签名的 canonical payload 字段
- HMAC-SHA256 signature
- 关键错误码样例，例如 `channel_origin_denied`、`channel_rate_limited`、`duplicate_inbound`

每个验签 profile 会声明平台侧 `signatureHeader`、`timestampHeader`、`replayHeader`、canonical payload 字段、样例 canonical payload、样例签名、重试语义和关键失败码。`scripts/check-channel-examples.js` 会在质量门禁里重新计算请求样例和平台 profile 的签名，并确认样例 endpoint、失败样例引用、矩阵 profile 引用和错误码仍存在于 API contract。

本地服务启动后，可以执行 `./scripts/demo-channel-inbound.sh`，脚本会读取同一份样例，动态刷新 timestamp、消息 ID 和签名，并把请求发送到运行中的渠道入口。

React 控制台的渠道接入区域也直接读取这份样例，展示 endpoint、来源头、secret ref、平台签名头、replay 头、签名预览、签名外部会话 ID 和预期状态码，并可用 demo secret 发送演示请求。请求成功后，控制台展示 Agent 回复、trace、证据标题和 fallback reason，便于课堂解释不同渠道如何归一到同一条可靠 Agent 链路。

控制台还会读取 `errorExamples`，提供来源不匹配、签名错误、过期 timestamp、重复消息和限流等失败演示，展示 `status + error.code`。Dashboard 会聚合 `ChannelAlert[]`、`ChannelFailureTrend[]`、`ChannelInboundAudit[]`、`ChannelAlertPolicy[]`、`ChannelNotification[]`、`ChannelRunbook[]`、`ChannelRunbookAssigneeLoad[]`、`ChannelRunbookCheckEvent[]`、`NotificationPolicyChange[]` 和 `NotificationPolicyEvent[]`，用于说明系统为什么不会盲收外部请求，以及失败如何进入验收审计、运营监控、通知策略、目标解析、secret ref 签名、delivery client、退避重试、外部回执、投递审计、死信、Runbook 处置和确认闭环；`ChannelInboundAudit` 只返回来源、结果、错误码、replay key、签名预览和 content hash，不返回完整 payload 或密钥值，并可通过 `/api/ops/channel-inbound-audits?channel=WeChat&status=REJECTED&code=signature` 追溯，或通过 `/api/ops/channel-inbound-audits/export` 导出 CSV。验收质量越线时会写入 `ChannelInboundAuditQualityEvent`，记录触发时的验收率、阈值快照和错误码，可通过 `/api/ops/channel-inbound-audit-quality-events?channel=WeChat&status=ESCALATE` 追溯，或通过 `/api/ops/channel-inbound-audit-quality-events/export` 导出 CSV。同一份数据也可通过 `/api/ops/channel-ops-report/export` 即时导出为 Markdown 或 CSV，或通过 `/api/ops/channel-ops-reports/generate` 生成 `ChannelOpsReport` 历史快照后下载；`ChannelOpsReport.summary.inboundAudit` 会记录 total、accepted、rejected、acceptanceRate 和 topErrorCodes；`summary.inboundAuditQuality` 会记录越线次数、ACTIVE/WATCH/RECOVERED 计数和对应渠道；`summary.runbookSummary` 会记录 Runbook 检查项 total/done/blocked/overdue/todo；`summary.runbookAssigneeLoads` 会记录负责人维度的 total/done/blocked/overdue/todo、涉及渠道和最近截止时间；`summary.runbookEvents` 会记录 ASSIGN/COMPLETE/BLOCK/RECOVER 计数和最近操作；`summary.handoffPriorities` 会按 Runbook 逾期、ACTIVE 验收越线、死信、高频失败和重试通知生成交接排序、原因、建议动作、`actionType/actionRef/actionLabel`、通知 ID 和 Runbook 状态锚点，控制台可从历史日报直接确认关联通知，并在确认备注写入 report id 与 action ref。Runbook 检查项可通过 `/api/ops/channel-runbook-checks/assign` 批量分派未完成步骤并生成 TODO 记录，通过 `/api/ops/channel-runbook-checks/complete` 确认完成，通过 `/api/ops/channel-runbook-checks/block` 标记阻塞，并通过 `/api/ops/channel-runbook-checks/recover` 恢复为完成；每次 ASSIGN/COMPLETE/BLOCK/RECOVER 都会写入 `ChannelRunbookCheckEvent`，可用 `/api/ops/channel-runbook-check-events?action=BLOCK&assignee=marketplace-oncall` 追溯，或通过对应 CSV 导出；`ChannelRunbook.checks` 会回显 action ref、step index、check status、处置摘要、负责人、截止时间、操作者和完成时间，未完成且 dueAt 已过期的检查项会进入 overdue 汇总和负责人负载统计，并可用 `/api/ops/channel-runbook-checks?assignee=marketplace-oncall&overdue=true` 或对应 CSV 导出筛选负责人和逾期项。`/api/ops/channel-ops-report-scheduler` 会返回后台调度开关、格式、间隔、保留数量、最近运行结果和下一次运行时间，控制台用它解释定时日报是否真的在工作；调度失败或运营需要补交接材料时，可通过 `/api/ops/channel-ops-report-scheduler/compensate` 立即补生成并写入 `ChannelOpsReportEvent`，随后用 `/api/ops/channel-ops-report-events?status=SUCCESS&actor=ops-lead` 追溯事件，或通过 `/api/ops/channel-ops-report-events/export` 导出 CSV 交接记录。`/api/ops/channel-alert-policies/update` 支持调整 target URL、secret ref、max attempts 和 backoff seconds；HIGH/CRITICAL 渠道会先生成 24 小时待审批变更，变更携带 current 快照、字段级 diff 和 `APPROVE <channel>` 确认短语，`approve-change` 仅允许 `ops-lead`、`security-owner`、`platform-owner` 携带确认短语后生效，`reject-change` 也要求审批人权限，`cancel-change` 允许申请方撤销，`rollback` 仅允许审批人携带 `ROLLBACK <channel>` 回滚到最近一次已批准变更的申请前快照，过期变更会在 Dashboard 和审批动作前自动标记为 `EXPIRED` 并记录审计。`ChannelRunbook` 会根据告警策略、通知状态、死信状态和按渠道配置的验收质量阈值给出下一步、升级条件和检查项；`ChannelNotification.deliveryAudit` 只返回 payload hash、签名预览和脱敏请求/响应摘要，不返回完整 signed payload 或密钥值。

`contracts/channel-protocol-matrix.json` 补充真实渠道差异表，说明各 adapter 的外部会话键、消息幂等键、客户字段、内容字段、时间字段、来源白名单、secret ref、平台签名头、时间头、replay key、rate limit、重试语义和关键失败码。控制台会把这张矩阵展示为“协议差异”，让字段映射、验签头和可靠性边界可以直接对照。

## Pagination

V1 数据量小，当前接口返回数组。后续列表增长后统一迁移到：

```json
{
  "records": [],
  "current": 1,
  "size": 20,
  "total": 100
}
```

## Verification

```bash
go test ./...
pnpm build:console
```
