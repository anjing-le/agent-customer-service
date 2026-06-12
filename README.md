# agent-customer-service

可靠智能客服教学项目，从 `infra-dev-scaffolding` 生长出来，聚焦多轮对话、RAG 检索增强、规则兜底、防幻觉、会话管理和历史记录。

底层技术栈、工程结构、契约和质量门禁沿用脚手架；本项目只新增客服 Agent 的业务设计。

## 当前边界

- 对话中心：`/api/chat/**`，会话、消息、多轮上下文、LLM/规则回复、推理过程。
- 知识中心：`/api/knowledge/**`，商品、活动、FAQ 参与当前对话检索。
- 场景配置：`/api/scene/**`，Intent/Prompt/Rule 管理、Prompt 测试，并轻量接入 Agent Runtime。

机器可读契约见 [contracts/service-boundaries.json](./contracts/service-boundaries.json)。

## 生长关系

```text
infra-dev-scaffolding
  ├─ contracts / scripts / project_document
  ├─ Vue + TypeScript + Vite
  └─ Spring Boot + Java 17

agent-customer-service
  ├─ chat: 会话与 Agent Runtime
  ├─ knowledge: 客服知识
  ├─ scene: 意图、提示词、规则
  └─ agent: 可靠性领域模型与编排
```

## 技术栈

- Frontend: Vue 3.5 + TypeScript + Vite 7 + Element Plus
- Backend: Spring Boot 3.4.5 + Java 17 + JPA
- Storage: MySQL 8；Redis/Redisson 可选
- LLM: OneRouter / OpenAI-compatible Chat Completions

## 快速开始

```bash
cd backend
mvn spring-boot:run
```

```bash
cd frontend
pnpm install
pnpm dev
```

后端默认端口 `10002`，前端默认端口 `20002`。本地敏感配置放在 `backend/src/main/resources/application-local.yml`，该文件已被 ignore，不应提交。

## 质量门禁

```bash
./scripts/check-template.sh
./scripts/check-contracts.sh
./scripts/quality-gate.sh
```

## 文档

- [项目约束](./project_document/PROJECT_CONSTRAINTS.md)
- [API 契约](./project_document/API_CONTRACT_GUIDE.md)
- [服务边界](./project_document/SERVICE_BOUNDARY_GUIDE.md)
- [脚手架继承](./project_document/SCAFFOLD_INHERITANCE.md)
- [领域模型](./project_document/DOMAIN_MODEL.md)
- [路线图](./project_document/ROADMAP.md)
- [本地启动](./project_document/LOCAL_STARTUP_GUIDE.md)
- [教学文档](./docs/teaching/00-环境准备与运行.md)

## License

MIT
