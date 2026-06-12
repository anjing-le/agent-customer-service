# Status

更新时间：2026-06-12

## 当前状态

| 领域 | 状态 | 说明 |
|---|---|---|
| 对话中心 | V2 runtime started | 会话/消息持久化仍在 `ChatService`，分析、检索、护栏、回复生成已由 `AgentRuntime` 编排 |
| 知识中心 | In progress | Product/Activity/FAQ 参与当前对话检索；Industry/Solution 为预留 |
| 场景配置 | Reserved runtime | CRUD 可用，Prompt 测试可调用 LLM，但未接入对话主链路 |
| 工程契约 | Ready for V1 | 已接入 contracts、scripts、project_document；前端 API 路径已集中到 `ApiPaths` |
| 响应 envelope | Ready for V1 | Chat/Knowledge/Scene Controller 已迁移到 `APIResponse<T>`；前端 HTTP 层兼容字符串成功码 `"0"` |
| Agent 领域模型 | In runtime | `ConversationTurn`、`KnowledgeRecall`、`GuardrailDecision`、`AgentReply` 已接入对话主链路 |
| OpenAPI typed client | Pending | 后续接入 `/v3/api-docs` 和前端 operation 类型 |
| 测试与质量门禁 | Ready for V1 | `./scripts/quality-gate.sh` 已通过；后续补单元测试和运行探针 |

## 当前保护边界

- 保留用户已有 README 收敛、OneRouter 地址更新和教学素材。
- 不读取或改动 ignored 的 `backend/src/main/resources/application-local.yml`。
- V2 第一阶段保持 API/前端 VO 不变，只替换后端内部编排结构。
- 业务运行链路优先收口，模板遗留接口后续按模块逐步清理。
- Scene 配置暂不影响运行时，下一阶段再接 Intent/Prompt/Rule。

## 验证入口

```bash
./scripts/check-template.sh
./scripts/check-contracts.sh
```

完整构建验证：

```bash
./scripts/quality-gate.sh
```
