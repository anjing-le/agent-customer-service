# agent-customer-service frontend

`agent-customer-service` 的 Vue 3 前端工程。

## 定位

前端提供可靠智能客服的运营和演示界面：

- 对话中心：会话、消息、知识选择、推理过程和知识召回。
- 知识中心：商品、活动、FAQ、行业知识、解决方案管理。
- 场景配置：意图、提示词模板、场景规则管理。

当前 V1 保留 `/api/chat/**`、`/api/knowledge/**`、`/api/scene/**` 运行路径。后续按根目录 `contracts/service-boundaries.json` 迁移到 `/api/customer-service/**`。

## 技术栈

- Vue 3.5 + TypeScript
- Vite 7
- Element Plus
- Pinia
- Axios
- Tailwind CSS 4

## 启动

```bash
pnpm install
pnpm dev
```

开发端口来自 `.env` 的 `VITE_PORT`，当前为 `20002`。开发环境通过 `VITE_API_PROXY_URL` 把 `/api` 转发到后端 `http://localhost:10002`。

## 构建

```bash
pnpm build
```

## 工程约束

- 页面不直接新增散落的 `/api/...`；新增接口先进入 API 层，后续迁移到 `ApiPaths` / OpenAPI typed client。
- 真实接口类型放在 `src/api/customer-service/model/**`。
- Product、Activity、FAQ 是 V1 真实参与对话检索的知识；Industry、Solution 是预留能力。
- Scene 配置当前未接入对话主链路，页面需要继续明确标注运行状态。
- UI 保持工作台工具气质，避免过重装饰和卡片套娃。

## 上游说明

前端工程基于 Art Design Pro 定制，保留 `LICENSE` 中的上游许可说明。
