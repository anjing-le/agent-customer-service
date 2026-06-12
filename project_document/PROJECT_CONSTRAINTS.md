# Project Constraints

本文档定义 `agent-customer-service` 的长期工程约束。项目必须保持从 `infra-dev-scaffolding` 生长出来的工程习惯，目标不是套模板，而是把“可靠 Agent 客服”做成清晰、可验证、可演进的工程系统。

## 总体边界

- 项目定位是智能客服：多轮会话、RAG 检索增强、规则兜底、防幻觉、会话管理和历史记录。
- 底层技术栈、目录习惯、契约、脚本和质量门禁继承脚手架；业务项目只扩展客服 Agent 设计。
- V1 保留当前单体架构，先把运行链路、接口契约和模块边界讲清楚。
- 不默认引入网关、注册中心、MQ、对象存储或链路追踪平台；重型能力只保留接口、配置样例和扩展点。
- Product、Activity、FAQ 是 V1 真实参与对话链路的知识；Industry、Solution 是预留能力，未接入前必须在页面和文档中标明。
- Scene 的 Intent、Prompt、Rule 已轻量接入运行时；完整条件表达式和 RuleEngine 未完成前，文档必须标明当前能力边界。

## 后端约束

- 新增 Controller 返回 `APIResponse<T>`；旧 `R<T>` 只作为迁移兼容层保留。
- 分页 payload 使用 `records/current/size/total`，逐步替换旧 `page/size/total`。
- Controller 路径必须引用 `ApiConstants`，后续按 `contracts/service-boundaries.json` 生成或校验。
- 真实请求和响应必须使用明确 DTO / VO，不使用 `Map<String, Object>` 承载主要业务 payload。
- 会话编排、知识检索、规则兜底、LLM 调用、防幻觉校验要分模块，不继续把逻辑堆进 `ChatService`。
- 业务当前时间逐步收口到 `DateUtils`，请求链路逐步补齐 `requestId`、`traceId`、语言和时区。
- LLM API Key、数据库密码等敏感配置只允许出现在本地 ignored 文件或环境变量中。

## 前端约束

- 页面和组件不直接拼 `/api/...`；V1 兼容路径集中在 `src/api/customer-service/**`，后续迁移到 `ApiPaths`/OpenAPI typed client。
- API 类型放在 `src/api/customer-service/model/**`；页面不写大段 `any`。
- 对话工作台优先服务客服运营流程：会话、消息、知识上下文、推理/召回证据。
- UI 遵守脚手架基线：克制文案、清晰密度、轻边界、少装饰，避免卡片套娃和过重圆角。
- 生产环境 API 不指向 Apifox mock。

## 质量门禁

- 第一阶段必须通过 `./scripts/check-template.sh` 和 `./scripts/check-contracts.sh`。
- 每批后端改动至少运行 `cd backend && mvn -q -DskipTests package`。
- 每批前端改动至少运行 `cd frontend && pnpm build`。
- 后续引入 OpenAPI 之后，新增接口必须同步生成前端类型并纳入脚本检查。
