# Production Extension Guide

## 定位

当前项目已经完成可靠 Agent 客服的教学样例：脚手架负责工程习惯，业务层只讲多轮会话、RAG evidence、无证据兜底、转人工、渠道治理和运营交接。

剩余约 4% 不应该继续堆进课堂主链路，而应该作为生产化扩展章节逐步展开。这样读者能看到项目如何从 DVSkyFolding 脚手架自然生长，而不是被新的基础设施复杂度盖住设计重点。

## 扩展原则

- 技术栈继续继承 DVSkyFolding：Go、React/Vite、PostgreSQL、contracts、scripts、env config、JSON logs、quality gate。
- 新能力优先加在已有端口后面，不绕过 `internal/customer`、`internal/knowledge`、`internal/ops` 和 `internal/platform/store` 的边界。
- 可靠客服边界不能弱化：有 evidence 才回答，无 evidence 就兜底，高风险转人工，所有关键动作要能审计。
- 生产化能力要先补契约、迁移、测试和控制台状态，再接外部平台或付费服务。
- 教学样例保持轻量，不提前引入向量数据库、权限系统、多租户网关或成本平台。

## 扩展章节

### 1. 向量检索和生产级 Rerank

当前教学版检索在 `Store.SearchKnowledge` 边界内完成 lightweight hybrid rerank，并返回 `retrievalScore` / `retrievalReason`。生产化时应该替换这个边界后的实现，而不是改写 Agent Runtime。

建议做法：

- 增加 embedding 生成任务和向量索引存储，可选 PostgreSQL pgvector 或独立向量数据库。
- 保留 `KnowledgeArticle.retrievalScore` 和 `retrievalReason`，让控制台仍能解释证据来源。
- 在 rerank adapter 内处理召回、重排、阈值和降级，外层仍只关心 evidence 是否可信。
- 回归用例继续覆盖知识命中、弱相关拒答、无证据兜底和多候选排序。

验收点：

- API contract 更新检索元数据。
- PostgreSQL migration 或外部 adapter 配置清晰。
- `pnpm verify` 和课堂 smoke 仍通过。
- 控制台能看见 score/reason，不暴露内部模型细节。

### 2. 权限和审批

当前样例已经展示通知策略审批、Runbook 操作审计和人工处置闭环。生产化需要把这些动作接入真实身份和权限。

建议做法：

- 所有运营写动作都带 `actor`，并由后端校验角色和权限。
- 区分客服坐席、质检、渠道运营、管理员和审批人。
- 高风险通知策略、规则发布、Runbook 批量分派必须有后端权限门禁。
- 控制台只展示可操作状态，不能依赖前端隐藏按钮作为安全边界。

验收点：

- contract 中声明 actor、role 和 forbidden response。
- 后端路由层返回统一 `forbidden` envelope。
- 审批、撤销、回滚和批量分派均保留审计事件。

### 3. 多租户

多租户会影响几乎所有数据对象，必须作为独立阶段推进。

建议做法：

- 为 conversation、message、knowledge、ticket、channel integration、report、runbook 等核心对象增加 `tenantId`。
- 所有 store query 默认按 tenant scope 过滤。
- inbound channel 根据 integration 或 app key 解析 tenant。
- 报表、审计和导出只能读取当前 tenant 的数据。

验收点：

- migration 增加 tenant 字段和必要索引。
- seed runtime 和 PostgreSQL runtime 行为一致。
- contract examples 覆盖 tenant scoped 请求。
- 测试验证跨 tenant 数据不可见。

### 4. 成本治理

模型生成、向量检索、渠道通知和日报任务都可能产生真实成本。生产化后需要把成本从一开始就纳入治理。

建议做法：

- 记录 LLM 调用、token、耗时、模型、fallback reason 和估算成本。
- 记录向量检索、rerank、通知投递和日报生成的次数与失败率。
- Dashboard 增加按渠道、租户和时间窗口聚合的成本摘要。
- 对高成本行为设置限流、预算阈值和告警。

验收点：

- 成本数据不影响主链路可靠性。
- 模型失败或成本超限时仍能回退到规则 RAG。
- 导出和日报能解释成本异常来源。

### 5. 真实渠道生产加固

当前样例已经覆盖多个渠道 adapter、签名、timestamp、replay、来源白名单、限流、external message id 和验收审计。生产化重点是补齐各平台真实差异。

建议做法：

- 按平台完善官方 message id、重试语义、回调字段、错误码和速率限制。
- 对接真实渠道前先补 contract examples 和验签样例集。
- 入站失败必须进入审计，不保存完整敏感 payload。
- 出站通知要保留签名、重试、死信、确认和人工补偿闭环。

验收点：

- 每个渠道有成功、失败、重放、签名错误和过期样例。
- Dashboard 可以解释验收率、主要失败码和 Runbook 处置状态。
- 日报能给出交接优先级建议。

### 6. 可观测和运维

教学样例已经有 JSON log、trace 字段、Dashboard 和质量门禁。生产化应继续沿用脚手架习惯，把运行状态变成可诊断的事实。

建议做法：

- 将 Agent trace、channel audit、notification delivery、report scheduler 和 runbook events 接入统一 metrics。
- 为核心链路增加 request id / conversation id / channel id / tenant id。
- 定义健康检查、告警阈值和排障 Runbook。
- 保留脚本化 quality gate，不让生产化接入绕开 `pnpm verify`。

验收点：

- 出问题时能从请求追到会话、证据、通知、Runbook 和日报。
- 关键失败都有可查询事件，不依赖控制台临时状态。

## 不放进当前教学样例的内容

- 不在课堂样例里强依赖向量数据库或生产 rerank 服务。
- 不提前引入完整认证网关、组织系统和多租户权限矩阵。
- 不把成本平台、告警平台和真实渠道 SDK 作为启动前置条件。
- 不为了演示“生产化”增加与 DVSkyFolding 技术栈无关的新框架。

## 推荐顺序

1. 替换检索实现：向量召回 + 生产 rerank，但保持 `SearchKnowledge` 边界稳定。
2. 接入权限和审批：先保护运营写动作，再保护配置和发布。
3. 推进多租户：先改数据模型和查询边界，再接租户级控制台。
4. 增加成本治理：先记录，再聚合，再设置预算和告警。
5. 加固真实渠道：按平台补齐官方协议差异和失败样例。
6. 强化可观测：把 trace、audit、scheduler、notification、runbook 串成可排障链路。

## 每个扩展章节的完成定义

- 契约更新：`contracts/api-contract.json` 能描述新增字段和错误边界。
- 数据更新：PostgreSQL migration 和 seed runtime 行为一致。
- 后端更新：能力落在已有模块边界或明确的新 adapter 中。
- 前端更新：控制台展示必要状态、空态、错误态和审计入口。
- 验证更新：单元测试、回归脚本、必要 smoke 和 `pnpm verify` 通过。
- 文档更新：README 或 project_document 说明这是脚手架上的业务增量。
