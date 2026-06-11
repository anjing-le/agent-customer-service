# Service Boundary Guide

`contracts/service-boundaries.json` 是 `agent-customer-service` 的服务边界事实来源。

## Current Boundaries

| boundary | kind | basePath | 说明 |
|---|---|---|---|
| chat | runtime | `/api/chat` | 会话、消息、多轮对话核心链路 |
| knowledge | runtime | `/api/knowledge` | 知识管理；Product/Activity/FAQ 参与当前对话检索 |
| scene | reserved-runtime | `/api/scene` | Intent/Prompt/Rule CRUD 和测试能力，当前未接入对话主链路 |

## Target Boundaries

后续统一迁移到 `/api/customer-service/**`：

- `/api/customer-service/chat/**`
- `/api/customer-service/knowledge/**`
- `/api/customer-service/scene/**`
- `/api/customer-service/ops/**`

## Runtime Truth

- 当前真实链路：用户消息 -> 分析 -> Product/Activity/FAQ 检索 -> LLM/规则回复 -> 推理过程返回。
- 当前预留链路：Industry/Solution 向量化、Scene 动态规则、Prompt runtime、RuleEngine。

## Update Rules

- 新增运行接口先更新 `contracts/service-boundaries.json`。
- 同步后端 `ApiConstants` 和前端 API 层。
- 运行 `./scripts/check-contracts.sh`。
