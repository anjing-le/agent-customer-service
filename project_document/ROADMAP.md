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
- 已落地机器可读 API contract，并纳入脚本校验。
- 已落地 Agent 回归验证集，并纳入 quality gate。

## V2: 可靠 Agent Runtime

- 已支持知识证据命中后回答。
- 已支持无证据时拒绝自由生成并创建知识缺口。
- 已支持知识缺口关闭和由缺口生成可信知识。
- 已支持规则测试：转人工、无证据兜底、可回答边界。
- 已支持转人工自动生成 ticket，并由控制台处理。
- 已支持人工接管时间线，记录创建、处理人、备注和解决事件。
- 已支持坐席 SLA、自动升级和人工工单筛选。
- 已支持会话历史详情和可选模型生成。
- 已支持 `Message.trace`，可观测策略、证据数、历史数、模型耗时和回退原因。
- 已支持批量回归验证：知识命中、无证据兜底、转人工、模型成功和模型失败回退。
- 已支持 SSE 流式回复，控制台通过 stream endpoint 展示增量输出。
- 已支持质量评估摘要和固定质量评估集，进入 quality gate。
- 已支持人工质检标注，按 groundedness、safety、helpfulness 维度给助手回复打分，并回写 Dashboard 质量摘要。
- 已支持渠道策略，Web/WeChat/WeCom/App/Marketplace/Douyin/Xiaohongshu 可配置语气、风险加权和转人工 SLA，并进入控制台筛选。
- 已支持标注复盘面板和低分样本导出，运营可导出 prompt/answer/评分/备注作为复盘材料。
- 已支持渠道 inbound 适配器，外部渠道消息通过 env 密钥、HMAC-SHA256 签名、timestamp 时间窗和 replay 记录后进入统一会话链路。
- 已支持渠道集成治理视图，展示 secret ref、签名窗口、replay 开关和轮换提示，但不暴露密钥值。
- 已支持 webhook 校验读取 `ChannelIntegration`，将 enabled、secret ref、签名窗口和 replay 开关接入运行时。
- 已支持 `externalMessageId` 对账，真实渠道可用官方消息 ID 做幂等，缺省时退回签名载荷 replay key。
- 已支持 WeChat/WeCom/App/Marketplace/Douyin/Xiaohongshu adapter，将真实渠道字段归一到标准 inbound 链路。
- 已支持 active/next secret ref 双密钥轮换窗口，控制台不暴露密钥值。
- 已支持渠道来源白名单，声明 `Origin` 或 `X-Channel-Origin` 时必须命中 `allowedOrigins`。
- 已支持渠道每分钟频率限制，超过 `rateLimitPerMinute` 返回 `channel_rate_limited`。
- 已支持渠道通知出站治理，包含通知目标配置、target URL、secret ref 签名、高风险变更审批/拒绝/撤销/过期、审批人权限、二次确认短语、字段级 diff、通知目标回滚、配置变更审计、退避重试、外部回执、死信、确认、脱敏投递审计和控制台筛选。
- 已支持真实平台回调协议差异和渠道验签样例集，包含 WeChat、WeCom、App、Marketplace、Douyin、Xiaohongshu 的平台签名头、时间头、replay 头、canonical payload、样例签名和重试语义。
- 已支持真实渠道验收审计，把成功/失败入站请求的来源、验签结果、幂等 key、签名预览、content hash 和失败原因沉淀到 Dashboard。
- 已支持渠道验收审计按渠道、状态和错误码筛选，并可导出 CSV 接入质量记录。
- 已支持渠道验收审计趋势摘要，渠道运营日报会展示 accepted/rejected、验收率和高频错误码。
- 已支持渠道验收趋势控制台可视化，按渠道展示 accepted/rejected 比例、验收率和高频错误码。
- 已支持渠道验收质量阈值配置化，按渠道定义最小样本数、最低验收率和同错误码触发门槛，异常时自动生成 Runbook 提醒。
- 已支持渠道验收质量越线事件，触发 Runbook 时沉淀验收率、阈值快照和错误码，并支持查询与 CSV 导出。
- 已支持渠道验收质量越线恢复视图，控制台可对照最近越线事件、当前验收率和渠道阈值判断 ACTIVE/WATCH/RECOVERED。
- 已支持渠道运营日报验收质量摘要，Markdown/CSV 和历史日报 summary 会展示越线次数、恢复渠道和仍需处理渠道。
- 已支持渠道运营日报交接优先级建议，把 ACTIVE 越线渠道、死信通知、高频失败码和重试通知排序成运营交接清单。
- 已支持交接优先级的确认锚点，每条建议会带 action type/ref/label、通知 ID 和 Runbook 状态，便于后续接入确认事件。
- 已支持历史日报交接通知确认动作，控制台可从日报卡片直接确认关联通知，并把 report id 与 action ref 写入确认备注。
- 已支持渠道失败告警 Runbook 和运营处置步骤，根据告警策略、通知状态和死信情况生成下一步、升级条件和检查项。
- 已支持渠道告警指标趋势图和控制台运营日报 JSON 导出，覆盖失败聚合、小时桶、通知状态、Runbook 和死信摘要。
- 已支持渠道运营日报的 Markdown/CSV 服务端导出，可供脚本、CI 或运营工具直接下载。
- 已支持渠道运营日报历史留存，控制台可生成 Markdown/CSV 快照、查看最近日报并下载历史文件。
- 已支持渠道运营日报定时任务触发器和保留策略，`platform-all` / `ops-api` 可通过 env 开启后台生成并保留最近 N 份快照。
- 已支持渠道运营日报调度状态可视化，控制台展示调度开关、格式、间隔、保留数量、最近运行结果和下一次运行时间。
- 已支持渠道运营日报失败时的人工补偿动作，控制台可立即补生成并记录 `ChannelOpsReportEvent` 审计事件。
- 已支持渠道运营日报补偿事件按状态/操作者筛选，并可导出 CSV 交接记录。
- 下一步为 Runbook 检查项增加可确认状态，把日报 action ref 关联到具体检查项的完成记录。

## V3: 生产化扩展

- 接入向量数据库和 rerank，替换 V1 关键词检索。
- 真实渠道接入生产化：Web、App、微信、抖音等渠道频率限制、回调协议差异和官方 message id 对账。
- 更完整的质检任务分配、规则版本灰度和复盘样本流转。
- 多租户、权限、限流、成本统计和治理面板。
- 继续扩展 DVSkyFolding 口径下的生产化运维、审计和治理能力。
