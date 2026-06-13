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

## V2: 可靠 Agent Runtime

- 已支持知识证据命中后回答。
- 已支持无证据时拒绝自由生成并创建知识缺口。
- 已支持知识缺口关闭和由缺口生成可信知识。
- 已支持规则测试：转人工、无证据兜底、可回答边界。
- 已支持转人工自动生成 ticket，并由控制台处理。
- 下一步补批量回归验证、会话历史详情和更完整的人工接管记录。

## V3: 生产化扩展

- 接入向量数据库和 rerank，替换 V1 关键词检索。
- 支持 SSE 流式回复。
- 多渠道适配：Web、App、微信、抖音等。
- 对话质量评估集、回归测试和人工标注反馈。
- 多租户、权限、限流、成本统计和治理面板。
- 删除旧运行入口和旧教学叙事，完成 DVSkyFolding 口径收敛。
