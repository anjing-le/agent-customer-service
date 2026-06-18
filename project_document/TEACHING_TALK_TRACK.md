# Teaching Talk Track

这份口播稿用于课堂、录屏和项目讲解。它的目标是把注意力放在设计，而不是让学生被底层工程细节淹没。

## 开场 30 秒

一句话：

> 这个项目不是重新造一套技术栈，而是展示一个智能客服业务如何从安静 DVSkyFolding 脚手架里长出来。

展开：

- 脚手架给我们 Go、React/Vite、PostgreSQL、contracts、scripts 和 quality gate。
- 业务只新增可靠 Agent 客服设计：多轮会话、RAG evidence、无证据兜底、规则转人工、渠道验收和运营交接。
- 学这个项目时，不需要重新学习一套工程习惯，只需要观察业务边界怎么设计。

## 1. 先讲脚手架基线

说法：

> 先看底座。这个项目沿用 DVSkyFolding 的目录、命令、配置、日志、数据库迁移和质量门禁。换业务时，这些习惯不变。

点哪里：

- `README.md`
- `apps/console`
- `cmd/platform-all`
- `internal/*`
- `contracts/`
- `scripts/`
- `infra/postgres/migrations`

证明什么：

- 前端是 React + TypeScript + Vite。
- 后端是 Go + `net/http` / `ServeMux`。
- 数据是 PostgreSQL + `pgx/v5` + SQL migrations。
- API 和服务边界有机器可读 contract。
- `pnpm verify` 能把模板、契约、Go 测试、Agent 回归、质量评估和 console build 一起跑完。

## 2. 再讲客服主链路

说法：

> 可靠客服 Agent 的第一原则不是会说，而是有证据才说。没有证据就兜底，高风险就转人工。

点哪里：

- 控制台“对话中心”
- RAG 证据区
- `Message.trace`

演示：

- 问“7 天无理由退货的运费怎么计算？”
- 看 evidence、`retrievalScore`、`retrievalReason` 和 trace。
- 再问知识库没有的问题，看知识缺口。
- 问投诉、催办或人工诉求，看转人工 ticket。

证明什么：

- RAG 是 evidence boundary，不是装饰。
- 可解释 rerank 能告诉学生“为什么召回这条知识”。
- 无证据时系统不会自由发挥。

## 3. 接着讲知识和规则运营

说法：

> Agent 不只是生成回复，还要把不确定、低质量和高风险样本交给运营闭环。

点哪里：

- 知识缺口
- 规则测试
- 质检任务
- 人工标注
- 复盘样本

证明什么：

- 知识缺口可以关闭，也可以生成可信知识。
- 规则可以测试、灰度、审批、发布和回滚。
- 低分样本能进入复盘材料，而不是只停留在一次回答里。

## 4. 然后讲渠道接入

说法：

> 真实渠道不能写成一堆散乱入口。不同平台先由 adapter 归一字段，再走同一套 inbound 校验。

点哪里：

- 渠道接入
- 协议样例
- 协议差异
- 失败演示

证明什么：

- Web、WeChat、WeCom、App、Marketplace、Douyin、Xiaohongshu 都归一到标准 inbound。
- 系统检查签名、来源、限流、timestamp、external message id 和 replay。
- 渠道请求不会盲收。

## 5. 最后讲验收、Runbook 和日报

说法：

> 渠道请求不是返回 200 就结束。生产里的关键是能审计、能恢复、能交接。

点哪里：

- 入站验收审计
- 验收质量越线
- 通知策略
- Runbook
- 负责人负载
- 日报留存

证明什么：

- `ChannelInboundAudit` 记录成功和失败，但不保存完整 payload 或密钥。
- 验收率过低或同错误码高频出现时，会派生 Runbook。
- `ChannelRunbookCheckEvent` 记录 ASSIGN、COMPLETE、BLOCK、RECOVER。
- Markdown/CSV 日报能把失败摘要、验收率、负责人负载和交接优先级带走。

## 收尾 30 秒

说法：

> 这就是这个样例要教的东西：底层工程习惯稳定复用，业务只关注可靠 Agent 的边界。当前教学样例已经收口，生产化扩展再进入向量检索、权限、多租户和成本治理。

必跑命令：

```bash
pnpm check:delivery
pnpm verify
pnpm demo:classroom:local
```

收尾判断：

- 能讲清脚手架继承。
- 能跑通客服主链路。
- 能解释 RAG evidence 和无证据兜底。
- 能展示渠道验收、Runbook 和日报交接。
- 能说明 V3 生产化扩展不压进当前教学样例。

## 讲课时不要偏

- 不要把它讲成模型能力展示；重点是可靠边界。
- 不要把向量数据库、多租户、鉴权、成本平台塞进当前课堂主线；这些是 V3。
- 不要重新解释一套工程技术选型；脚手架已经负责这件事。
- 不要只讲前端页面；每个页面都要对应 API、contract、测试或审计事实。
