# Domain Model

本文档描述 `agent-customer-service` 的可靠客服 Agent 领域模型。当前运行链路仍由 `ChatService` 承载，新模型先作为拆分边界和后续实现端口。

## Core Concepts

| 模型 | 说明 | 当前来源 |
|---|---|---|
| `ConversationTurn` | 一次用户输入触发的完整处理回合 | `ChatDTO.SendMessageDTO`、`ChatMessage`、会话上下文 |
| `ConversationMessage` | 会话历史中的单条消息 | `ChatMessage` |
| `IntentAnalysis` | 场景、意图、置信度、情绪和分析引擎 | `LlmService.analyzeUserInput` 或关键词兜底 |
| `KnowledgeRecall` | RAG 或关键词检索返回的知识证据集合 | Product、Activity、FAQ 检索结果 |
| `KnowledgeEvidence` | 可被回复引用的一条证据 | 商品、活动、FAQ、行业、解决方案、人工选择 |
| `GuardrailDecision` | 防幻觉、安全和兜底决策 | 规则兜底、低置信度、无可靠知识、LLM 不可用 |
| `AgentReply` | 一轮回复的完整结果 | `ChatVO.SendMessageVO` |
| `ReasoningStep` | 可审计推理过程 | `ChatVO.ReasoningStepVO` |

## Application Ports

后端新增 `com.anjing.module.agent.application` 端口，后续按端口拆出实现：

| 端口 | 职责 | 不应该做 |
|---|---|---|
| `ConversationMemory` | 读取最近历史、写入用户消息和上下文 | 不调用 LLM，不做检索 |
| `IntentAnalyzer` | 识别场景、意图、情绪和置信度 | 不生成最终回复 |
| `KnowledgeRetriever` | 按意图和上下文召回知识证据 | 不拼接最终话术 |
| `GuardrailPolicy` | 判断是否安全、是否要兜底、兜底原因 | 不直接访问 Controller |
| `ReplyGenerator` | 基于证据生成 LLM 或规则回复 | 不保存会话 |
| `AgentRuntime` | 编排一轮 Agent 处理 | 不承载具体仓储细节 |

## Runtime Boundary

当前真实链路：

```mermaid
flowchart LR
    A["User Message"] --> B["ChatService"]
    B --> C["IntentAnalysis: LLM or keyword"]
    C --> D["KnowledgeRecall: Product, Activity, FAQ"]
    D --> E["ReplyGenerator: LLM or rule fallback"]
    E --> F["ReasoningStep audit"]
    F --> G["ChatVO.SendMessageVO"]
```

目标链路：

```mermaid
flowchart LR
    A["ConversationTurn"] --> B["ConversationMemory"]
    A --> C["IntentAnalyzer"]
    C --> D["KnowledgeRetriever"]
    D --> E["GuardrailPolicy"]
    E --> F["ReplyGenerator"]
    F --> G["AgentReply"]
```

## Reliability Rules

- 回复必须能说明来源：`AgentReply.knowledgeRecall` 保留召回证据。
- 无可靠知识时不能编造：`GuardrailDecision.fallbackRequired` 应触发规则兜底或转人工。
- 低置信度意图不能强行执行动作：通过 `FallbackReason.LOW_CONFIDENCE` 标记。
- LLM 不可用不影响基础客服：通过 `FallbackReason.LLM_UNAVAILABLE` 使用规则回复。
- 安全或政策拦截必须可审计：通过 `policyTags` 和 `userVisibleNotice` 留痕。

## Migration Plan

1. 保持 `ChatService` 现有行为不变，将内部中间结果映射到领域对象。
2. 抽出 `IntentAnalyzer` 实现，复用当前 LLM 分析和关键词兜底。
3. 抽出 `KnowledgeRetriever` 实现，先保留关键词检索，后续替换为向量检索和 rerank。
4. 增加 `GuardrailPolicy`，将低置信度、无知识、敏感场景和转人工策略集中治理。
5. 最后引入 `AgentRuntime` 编排，Controller 只调用 runtime 并返回 `AgentReply` 映射后的 VO。
