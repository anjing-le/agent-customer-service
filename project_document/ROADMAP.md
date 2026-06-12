# Roadmap

## 定位

`agent-customer-service` 是可靠 Agent 客服项目，核心关注多轮对话、RAG 检索增强、规则兜底、防幻觉、会话管理和历史记录。

## V1: 工程对齐与可靠 Demo

- 补齐脚手架工程骨架：contracts、scripts、project_document、质量门禁入口。
- 保留现有对话中心、知识中心、场景配置三大页面。
- 将当前真实/模拟边界写入 README、文档和 service boundary。
- 统一项目身份、端口、环境说明和生产 API 地址。
- 逐步将旧 `R<T>`、手写 URL、漂移 TS 类型迁移到脚手架契约。
- 落地 Agent 领域模型和 application ports，先固定拆分边界，不破坏现有 Demo。

## V2: 可靠 Agent Runtime

- 已开始拆分 `ChatService`：会话/消息保留在 `ChatService`，分析、检索、生成、兜底、防幻觉、审计由 `AgentRuntime` 编排。
- Scene 配置接入运行时：动态 Intent、PromptRuntime、RuleEngine。
- 已将 `ConversationTurn`、`KnowledgeRecall`、`GuardrailDecision`、`AgentReply` 接入真实 runtime。
- 增加会话历史、可靠性看板和兜底原因追踪。
- 支持 SSE 流式回复和低置信度转人工策略。

## V3: 生产化扩展

- 接入向量数据库和 rerank，替换 V1 关键词检索。
- 多渠道适配：Web、App、微信、抖音等。
- 对话质量评估集、回归测试和人工标注反馈。
- 接入未来 `infra-llm-gateway`、`infra-auth`、`infra-api-gateway`。
- 多租户、权限、限流、成本统计和治理面板。
