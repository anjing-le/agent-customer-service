# Scaffold Inheritance

`agent-customer-service` 不是从零搭出的孤立项目，而是从 `infra-dev-scaffolding` 生长出来的可靠 Agent 客服样例。

教学目标是让学习者把注意力放在业务设计上：多轮会话、RAG、规则兜底、防幻觉、运行时可观测。底层工程习惯、契约、质量门禁和前后端技术栈由脚手架统一提供。

## Inherited From Scaffold

| 能力 | 脚手架约定 | 本项目状态 |
|---|---|---|
| 技术栈 | Vue 3.5 + TypeScript + Vite 7；Spring Boot 3.4.5 + Java 17 | 已继承，业务模块按客服场景扩展 |
| 项目结构 | `frontend/`、`backend/`、`contracts/`、`project_document/`、`scripts/` | 已对齐 |
| API 契约 | `/api` 前缀、统一 response envelope、分页约定、服务边界 manifest | 已有 `contracts/platform-contract.json` 和 `contracts/service-boundaries.json` |
| 质量门禁 | `check-template`、`check-contracts`、`quality-gate` | 已接入并通过 |
| 文档习惯 | 状态、路线图、边界、启动、约束分开维护 | 已接入项目文档 |
| 前端习惯 | API 路径集中、页面只消费业务数据、生产环境不指向 mock | 已收口到 `ApiPaths` 和 customer-service API |
| 后端习惯 | Controller 使用常量路径、统一响应、业务分层 | Chat/Knowledge/Scene 已迁移到 `APIResponse<T>`，Agent Runtime 已拆分 |

## Business Added By This Agent

脚手架不关心客服业务，`agent-customer-service` 只在业务层新增：

- `chat`: 会话、消息、多轮历史和 Agent Runtime 入口。
- `knowledge`: Product、Activity、FAQ 等客服知识管理和检索素材。
- `scene`: Intent、Prompt、Rule 配置，并轻量接入运行时。
- `agent.domain`: `ConversationTurn`、`KnowledgeRecall`、`GuardrailDecision`、`AgentReply` 等领域模型。
- `agent.runtime`: 分析、检索、护栏、回复生成和推理过程编排。

## Teaching Narrative

教学时按这个顺序讲：

1. 先看脚手架：目录、技术栈、契约、质量门禁。
2. 再看边界：客服 Agent 只新增 `chat/knowledge/scene/agent`。
3. 再看链路：用户消息如何进入 `AgentRuntime`。
4. 再看可靠性：知识证据、兜底原因、护栏标签如何被展示。
5. 最后看扩展：向量检索、完整 RuleEngine、PromptRuntime 如何继续生长。

## Non Goals

- 不在业务项目里重新发明脚手架已有的通用能力。
- 不为了单个 Agent 引入和脚手架不一致的前端框架、响应格式或脚本习惯。
- 不把教学重点放在模板复制，而是放在“业务如何沿着脚手架边界生长”。

## Verification

每批变更至少运行：

```bash
./scripts/check-template.sh
./scripts/check-contracts.sh
```

涉及前后端代码时运行：

```bash
./scripts/quality-gate.sh
```
