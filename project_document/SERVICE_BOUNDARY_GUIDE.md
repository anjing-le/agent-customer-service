# Service Boundary Guide

`contracts/service-boundaries.json` 是 `agent-customer-service` 的服务边界事实来源。

## Current Boundaries

| boundary | kind | basePath | 说明 |
|---|---|---|---|
| chat | runtime | `/api/chat` | 会话、消息、多轮对话核心链路 |
| knowledge | runtime | `/api/knowledge` | 知识管理；Product/Activity/FAQ 参与当前对话检索 |
| scene | runtime | `/api/scene` | Intent/Prompt/Rule CRUD 和测试能力；已轻量接入 Agent Runtime |

## Target Boundaries

后续统一迁移到 `/api/customer-service/**`：

- `/api/customer-service/chat/**`
- `/api/customer-service/knowledge/**`
- `/api/customer-service/scene/**`
- `/api/customer-service/ops/**`

## Runtime Truth

- 当前真实链路：用户消息 -> `AgentRuntime` -> 意图分析 -> Product/Activity/FAQ 检索 -> 护栏决策 -> LLM/规则回复 -> 推理和可靠性结果返回。
- Scene 当前接入：启用 Intent 参与关键词兜底识别；启用 SYSTEM Prompt 由 PromptRuntime 渲染后注入 LLM 上下文；启用 Rule 由 RuleEngine 产生命中原因和动作，支持内置规则码和轻量 JSON 条件表达式。
- 当前预留链路：Industry/Solution 向量化、Rule 条件编辑器校验、Prompt 变量 schema、向量检索和 rerank。

## Update Rules

- 新增运行接口先更新 `contracts/service-boundaries.json`。
- 同步后端 `ApiConstants` 和前端 API 层。
- 运行 `./scripts/check-contracts.sh`。
