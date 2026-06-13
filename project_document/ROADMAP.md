# Roadmap

## 定位

`agent-customer-service` 是可靠 Agent 客服项目，核心关注多轮对话、RAG 检索增强、规则兜底、防幻觉、会话管理、知识运营、人工接管和历史记录。

它是安静 DVSkyFolding 的业务生长样例：脚手架负责 Go/React/PostgreSQL/交付习惯，项目本身只讲客服 Agent 的领域设计。

## V1: DVSkyFolding 基线与可靠 Demo

- 已落地 Go module、React/Vite console、PostgreSQL migrations。
- 已落地多 command：`platform-all`、`customer-service-api`、`ops-api`、`console-web`、`migrate-db`。
- 已落地 seed runtime 和 PostgreSQL runtime。
- 已落地对话中心、知识中心、场景配置、人工队列的 React 控制台。
- 已落地统一 JSON envelope、结构化日志、env 配置。
- 已落地本地 PostgreSQL compose 和可选集成测试。
- 已落地机器可读 API contract，并纳入脚本校验。
- 已落地 Agent 回归验证集，并纳入 quality gate。

## V2: 可靠 Agent Runtime

- 已支持知识证据命中后回答。
- 已支持无证据时拒绝自由生成并创建知识缺口。
- 已支持知识缺口关闭和由缺口生成可信知识。
- 已支持规则测试：转人工、无证据兜底、可回答边界。
- 已支持转人工自动生成 ticket，并由控制台处理。
- 已支持人工接管时间线，记录创建、处理人、备注和解决事件。
- 已支持坐席 SLA、自动升级和人工工单筛选。
- 已支持会话历史详情和可选模型生成。
- 已支持 `Message.trace`，可观测策略、证据数、历史数、模型耗时和回退原因。
- 已支持批量回归验证：知识命中、无证据兜底、转人工、模型成功和模型失败回退。
- 已支持 SSE 流式回复，控制台通过 stream endpoint 展示增量输出。
- 已支持质量评估摘要和固定质量评估集，进入 quality gate。
- 已支持人工质检标注，按 groundedness、safety、helpfulness 维度给助手回复打分，并回写 Dashboard 质量摘要。
- 已支持渠道策略，Web/WeChat/App/Marketplace 可配置语气、风险加权和转人工 SLA，并进入控制台筛选。
- 已支持标注复盘面板和低分样本导出，运营可导出 prompt/answer/评分/备注作为复盘材料。
- 下一步补真实渠道接入适配器和渠道消息签名校验。

## V3: 生产化扩展

- 接入向量数据库和 rerank，替换 V1 关键词检索。
- 真实渠道接入适配器：Web、App、微信、抖音等消息入口、签名校验和回调。
- 更完整的质检任务分配、规则版本灰度和复盘样本流转。
- 多租户、权限、限流、成本统计和治理面板。
- 继续扩展 DVSkyFolding 口径下的生产化运维、审计和治理能力。
