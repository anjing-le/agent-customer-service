import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  AlertTriangle,
  BookOpen,
  Bot,
  CheckCircle2,
  Database,
  FileSearch,
  FilePlus2,
  MessageSquareText,
  RefreshCcw,
  ShieldCheck,
  UserRoundCheck
} from 'lucide-react';
import './styles.css';

type ApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
};

type Metric = { label: string; value: string; note: string };
type Conversation = {
  id: string;
  customer: string;
  channel: string;
  intent: string;
  status: string;
  riskLevel: string;
  lastMessage: string;
};
type KnowledgeGap = {
  id: string;
  question: string;
  reason: string;
  status: string;
  priority: string;
};
type Rule = { id: string; code: string; name: string; trigger: string; action: string; enabled: boolean };
type RuleTestResult = {
  input: string;
  matched: boolean;
  ruleCode?: string;
  action: string;
  riskLevel: string;
  fallback: boolean;
  reason: string;
  recommended: string;
};
type TransferTicket = {
  id: string;
  conversationId: string;
  question: string;
  reason: string;
  priority: string;
  status: string;
  slaMinutes: number;
  waitMinutes: number;
  slaStatus: string;
  escalated: boolean;
  assignee?: string;
  resolutionNote?: string;
  createdAt: string;
  resolvedAt?: string;
  events?: TransferEvent[];
};
type TransferEvent = {
  type: string;
  actor: string;
  note: string;
  createdAt: string;
};
type KnowledgeArticle = {
  id: string;
  title: string;
  category: string;
  content?: string;
  tags?: string[];
  trustLevel: string;
};
type AgentTrace = {
  strategy: string;
  evidenceCount: number;
  historyCount: number;
  modelAttempted: boolean;
  model?: string;
  modelDurationMs?: number;
  modelFallback: boolean;
  modelFallbackReason?: string;
};
type Message = {
  id: string;
  conversationId: string;
  role: string;
  content: string;
  engine: string;
  safe: boolean;
  fallbackReason?: string;
  evidenceIds?: string[];
  createdAt: string;
  trace?: AgentTrace;
};
type QualitySummary = {
  score: number;
  reviewedMessages: number;
  evidenceAnswers: number;
  safeFallbacks: number;
  humanTransfers: number;
  notes: string[];
};
type Dashboard = {
  metrics: Metric[];
  conversations: Conversation[] | null;
  knowledgeGaps: KnowledgeGap[] | null;
  rules: Rule[] | null;
  transfers: TransferTicket[] | null;
  quality: QualitySummary;
};
type SendMessageResult = {
  conversation: Conversation;
  agentMessage: Message;
  evidence: KnowledgeArticle[];
  gap?: KnowledgeGap;
};
type StreamEvent = {
  type: 'meta' | 'delta' | 'done';
  data?: unknown;
};

const api = async <T,>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) }
  });
  const payload = (await response.json()) as ApiResponse<T>;
  if (!payload.success || payload.data === undefined) {
    throw new Error(payload.error?.message ?? 'request failed');
  }
  return payload.data;
};

const traceChips = (trace?: AgentTrace) => {
  if (!trace) {
    return [];
  }
  const chips = [
    trace.strategy,
    `${trace.evidenceCount} evidence`,
    `${trace.historyCount} history`
  ];
  if (trace.modelAttempted) {
    chips.push(trace.model ? `model ${trace.model}` : 'model attempted');
    chips.push(`${trace.modelDurationMs ?? 0} ms`);
  }
  if (trace.modelFallback) {
    chips.push(trace.modelFallbackReason ? `fallback ${trace.modelFallbackReason}` : 'model fallback');
  }
  return chips;
};

const parseSSEBlock = (block: string): StreamEvent | null => {
  const dataLine = block.split('\n').find((line) => line.startsWith('data: '));
  if (!dataLine) {
    return null;
  }
  return JSON.parse(dataLine.slice(6)) as StreamEvent;
};

const postMessageStream = async (
  conversationId: string | undefined,
  content: string,
  onDelta: (content: string) => void
) => {
  const response = await fetch('/api/customer-service/messages/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ conversationId, content })
  });
  if (!response.ok || !response.body) {
    throw new Error(`stream failed: ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let finalResult: SendMessageResult | null = null;

  while (true) {
    const { value, done } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    const blocks = buffer.split('\n\n');
    buffer = blocks.pop() ?? '';
    for (const block of blocks) {
      const event = parseSSEBlock(block);
      if (!event) {
        continue;
      }
      if (event.type === 'delta') {
        const data = event.data as { content?: string };
        onDelta(data.content ?? '');
      }
      if (event.type === 'done') {
        finalResult = event.data as SendMessageResult;
      }
    }
  }

  if (!finalResult) {
    throw new Error('stream ended without final result');
  }
  return finalResult;
};

function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [knowledge, setKnowledge] = useState<KnowledgeArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState('7 天无理由退货的运费怎么计算？');
  const [ruleInput, setRuleInput] = useState('我已经投诉很多次了，现在必须转人工');
  const [ruleResult, setRuleResult] = useState<RuleTestResult | null>(null);
  const [result, setResult] = useState<SendMessageResult | null>(null);
  const [selectedConversationId, setSelectedConversationId] = useState('');
  const [history, setHistory] = useState<Message[]>([]);
  const [error, setError] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [streamReply, setStreamReply] = useState('');
  const [transferFilter, setTransferFilter] = useState<'ALL' | 'OPEN' | 'ESCALATED' | 'RESOLVED'>('ALL');

  const conversations = dashboard?.conversations ?? [];
  const gaps = dashboard?.knowledgeGaps ?? [];
  const rules = dashboard?.rules ?? [];
  const transfers = dashboard?.transfers ?? [];
  const visibleTransfers = transfers.filter((ticket) => {
    if (transferFilter === 'OPEN') {
      return ticket.status === 'OPEN';
    }
    if (transferFilter === 'ESCALATED') {
      return ticket.escalated;
    }
    if (transferFilter === 'RESOLVED') {
      return ticket.status === 'RESOLVED';
    }
    return true;
  });
  const activeConversation = useMemo(
    () => conversations.find((item) => item.id === selectedConversationId) ?? conversations[0],
    [conversations, selectedConversationId]
  );
  const highRiskCount = conversations.filter((item) => item.riskLevel === 'HIGH').length;
  const openGapCount = gaps.filter((item) => item.status === 'OPEN').length;
  const openTransferCount = transfers.filter((item) => item.status === 'OPEN').length;

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [dashboardData, knowledgeData] = await Promise.all([
        api<Dashboard>('/api/ops/dashboard'),
        api<KnowledgeArticle[]>('/api/knowledge/articles')
      ]);
      setDashboard(dashboardData);
      setKnowledge(knowledgeData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'load failed');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    if (!selectedConversationId && conversations[0]?.id) {
      setSelectedConversationId(conversations[0].id);
    }
  }, [conversations, selectedConversationId]);

  const loadMessages = async (conversationId: string) => {
    if (!conversationId) {
      setHistory([]);
      return;
    }
    setHistory(await api<Message[]>(`/api/customer-service/messages?conversationId=${encodeURIComponent(conversationId)}`));
  };

  useEffect(() => {
    if (!activeConversation?.id) {
      setHistory([]);
      return;
    }
    void loadMessages(activeConversation.id).catch((err) => {
      setError(err instanceof Error ? err.message : 'load messages failed');
    });
  }, [activeConversation?.id]);

  const send = async () => {
    setError('');
    setStreaming(true);
    setStreamReply('');
    try {
      const data = await postMessageStream(activeConversation?.id, message, (chunk) => {
        setStreamReply((current) => current + chunk);
      });
      setResult(data);
      setSelectedConversationId(data.conversation.id);
      await load();
      await loadMessages(data.conversation.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'send failed');
    } finally {
      setStreaming(false);
    }
  };

  const resolveGap = async (gap: KnowledgeGap) => {
    setError('');
    try {
      await api<KnowledgeGap>('/api/knowledge/gaps/resolve', {
        method: 'POST',
        body: JSON.stringify({ id: gap.id })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'resolve gap failed');
    }
  };

  const createArticleFromGap = async (gap: KnowledgeGap) => {
    setError('');
    try {
      await api<KnowledgeArticle>('/api/knowledge/gaps/create-article', {
        method: 'POST',
        body: JSON.stringify({
          gapId: gap.id,
          title: gap.question,
          category: '运营补充',
          content: `${gap.question}：请按最新客服政策补充标准答案。`,
          tags: [gap.reason, '运营补充']
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'create article failed');
    }
  };

  const testRule = async () => {
    setError('');
    try {
      setRuleResult(await api<RuleTestResult>('/api/ops/rules/test', {
        method: 'POST',
        body: JSON.stringify({ content: ruleInput })
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'test rule failed');
    }
  };

  const resolveTransfer = async (ticket: TransferTicket) => {
    setError('');
    try {
      await api<TransferTicket>('/api/ops/transfers/resolve', {
        method: 'POST',
        body: JSON.stringify({ id: ticket.id, assignee: 'operator', note: '人工已接管并回写处理结果' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'resolve transfer failed');
    }
  };

  return (
    <main className="shell">
      <aside className="nav">
        <div className="brand">
          <Bot size={24} />
          <span>Agent CS</span>
        </div>
        <a className="active"><MessageSquareText size={18} /> 对话中心</a>
        <a><BookOpen size={18} /> 知识中心</a>
        <a><ShieldCheck size={18} /> 场景配置</a>
        <a><UserRoundCheck size={18} /> 人工队列</a>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">DVSkyFolding Runtime</p>
            <h1>可靠客服 Agent 控制台</h1>
          </div>
          <div className="topActions">
            <span className="runtime"><Database size={16} /> Go · React · PostgreSQL</span>
            <button className="iconButton" onClick={load} aria-label="刷新" title="刷新">
              <RefreshCcw size={18} />
            </button>
          </div>
        </header>

        {error && <div className="notice"><AlertTriangle size={18} /> {error}</div>}

        <section className="metrics">
          {(dashboard?.metrics ?? []).map((item) => (
            <article className="metric" key={item.label}>
              <span>{item.label}</span>
              <strong>{item.value}</strong>
              <small>{item.note}</small>
            </article>
          ))}
          <article className="metric">
            <span>High risk</span>
            <strong>{highRiskCount}</strong>
            <small>sessions needing operator attention</small>
          </article>
          <article className="metric">
            <span>Human queue</span>
            <strong>{openTransferCount}</strong>
            <small>open transfer tickets</small>
          </article>
          {loading && <article className="metric skeleton">Loading</article>}
        </section>

        {dashboard?.quality && (
          <section className="qualityBand">
            <div>
              <p className="sectionLabel">质量评估</p>
              <strong>{dashboard.quality.score}</strong>
              <span>{dashboard.quality.reviewedMessages} reviewed messages</span>
            </div>
            <div className="qualityStats">
              <span>{dashboard.quality.evidenceAnswers} evidence</span>
              <span>{dashboard.quality.safeFallbacks} fallback</span>
              <span>{dashboard.quality.humanTransfers} transfer</span>
            </div>
            <div className="qualityNotes">
              {dashboard.quality.notes.map((note) => <small key={note}>{note}</small>)}
            </div>
          </section>
        )}

        <section className="runtimeGrid">
          <section className="panel agentPanel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">对话中心</p>
                <h2>Agent 工作区</h2>
              </div>
              <span className={statusClass(activeConversation?.riskLevel)}>{activeConversation?.status ?? 'Ready'}</span>
            </div>
            <div className="activeSession">
              <strong>{activeConversation?.customer ?? '访客'}</strong>
              <span>{activeConversation?.intent ?? '待识别'} · {activeConversation?.channel ?? 'Web'}</span>
            </div>
            <textarea value={message} onChange={(event) => setMessage(event.target.value)} />
            <div className="actionRow">
              <button className="primary" onClick={send} disabled={streaming}>{streaming ? '流式生成中' : '发送并审计'}</button>
              <span>{streaming ? 'streaming reply' : `${openGapCount} open gaps`}</span>
            </div>
            {streaming && (
              <div className="reply streamingReply">
                <div className="replyMeta">
                  <span>SSE</span>
                  <span>streaming</span>
                </div>
                <p>{streamReply || '等待 Agent 回复...'}</p>
              </div>
            )}
            {result && (
              <div className="reply">
                <div className="replyMeta">
                  <span>{result.agentMessage.engine}</span>
                  <span>{result.agentMessage.fallbackReason ?? 'EVIDENCE_OK'}</span>
                  <span>{result.evidence.length} evidence</span>
                </div>
                <div className="traceMeta">
                  {traceChips(result.agentMessage.trace).map((item) => <small key={item}>{item}</small>)}
                </div>
                <p>{result.agentMessage.content}</p>
              </div>
            )}
            <div className="history">
              <div className="historyHeader">
                <strong>会话历史</strong>
                <span>{history.length} messages</span>
              </div>
              {history.map((item) => (
                <article className={`message ${item.role}`} key={item.id}>
                  <span>{item.role} · {item.engine}</span>
                  <p>{item.content}</p>
                  <div className="traceMeta">
                    {traceChips(item.trace).map((trace) => <small key={trace}>{trace}</small>)}
                  </div>
                  {(item.evidenceIds?.length ?? 0) > 0 && <small>{item.evidenceIds?.join(', ')}</small>}
                  {item.fallbackReason && <small>{item.fallbackReason}</small>}
                </article>
              ))}
              {history.length === 0 && <p className="empty">选择会话后展示历史消息</p>}
            </div>
          </section>

          <section className="panel queuePanel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">会话运行</p>
                <h2>会话队列</h2>
              </div>
            </div>
            <div className="tableList">
              {conversations.map((item) => (
                <article
                  className={`tableRow selectable ${item.id === activeConversation?.id ? 'selected' : ''}`}
                  key={item.id}
                  onClick={() => setSelectedConversationId(item.id)}
                >
                  <div>
                    <strong>{item.customer}</strong>
                    <span>{item.lastMessage}</span>
                  </div>
                  <div>
                    <em>{item.intent}</em>
                    <b className={statusClass(item.riskLevel)}>{item.status}</b>
                  </div>
                </article>
              ))}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">人工队列</p>
                <h2>待接管工单</h2>
              </div>
              <span className="status">{openTransferCount}</span>
            </div>
            <div className="filterRow">
              {(['ALL', 'OPEN', 'ESCALATED', 'RESOLVED'] as const).map((item) => (
                <button
                  className={transferFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => setTransferFilter(item)}
                >
                  {item}
                </button>
              ))}
            </div>
            <div className="tableList">
              {visibleTransfers.map((ticket) => (
                <article className="tableRow" key={ticket.id}>
                  <div>
                    <strong>{ticket.question}</strong>
                    <span>{ticket.reason} · {ticket.conversationId}</span>
                    <span>{ticket.slaStatus} · {ticket.waitMinutes}m / {ticket.slaMinutes}m SLA</span>
                    {ticket.assignee && <span>{ticket.assignee} · {ticket.resolutionNote}</span>}
                    <div className="timeline">
                      {(ticket.events ?? []).map((event) => (
                        <small key={`${ticket.id}-${event.type}-${event.createdAt}`}>
                          {event.type} · {event.actor} · {event.note}
                        </small>
                      ))}
                    </div>
                  </div>
                  <div className="gapActions">
                    <b className={statusClass(ticket.priority)}>{ticket.escalated ? 'ESCALATED' : ticket.status}</b>
                    {ticket.status === 'OPEN' && (
                      <span>
                        <button className="tinyButton" onClick={() => resolveTransfer(ticket)} title="处理工单">
                          <UserRoundCheck size={14} />
                        </button>
                      </span>
                    )}
                  </div>
                </article>
              ))}
              {visibleTransfers.length === 0 && <p className="empty">暂无匹配工单</p>}
            </div>
          </section>

          <section className="panel evidencePanel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">RAG 证据</p>
                <h2>本轮召回</h2>
              </div>
              <FileSearch size={18} />
            </div>
            <div className="evidenceList">
              {(result?.evidence.length ? result.evidence : knowledge.slice(0, 3)).map((item) => (
                <article className="evidence" key={item.id}>
                  <div>
                    <strong>{item.title}</strong>
                    <span>{item.category} · {item.trustLevel}</span>
                  </div>
                  {item.content && <p>{item.content}</p>}
                </article>
              ))}
            </div>
          </section>

          <section className="panel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">知识中心</p>
                <h2>知识缺口</h2>
              </div>
              <span className="status">{openGapCount}</span>
            </div>
            <div className="tableList">
              {gaps.map((item) => (
                <article className="tableRow" key={item.id}>
                  <div>
                    <strong>{item.question}</strong>
                    <span>{item.reason}</span>
                  </div>
                  <div className="gapActions">
                    <b className={statusClass(item.priority)}>{item.status}</b>
                    {item.status === 'OPEN' && (
                      <span>
                        <button className="tinyButton" onClick={() => createArticleFromGap(item)} title="生成知识">
                          <FilePlus2 size={14} />
                        </button>
                        <button className="tinyButton" onClick={() => resolveGap(item)} title="关闭缺口">
                          <CheckCircle2 size={14} />
                        </button>
                      </span>
                    )}
                  </div>
                </article>
              ))}
              {gaps.length === 0 && <p className="empty">暂无开放缺口</p>}
            </div>
          </section>

          <section className="panel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">场景配置</p>
                <h2>规则测试</h2>
              </div>
              <ShieldCheck size={18} />
            </div>
            <textarea className="compactInput" value={ruleInput} onChange={(event) => setRuleInput(event.target.value)} />
            <div className="actionRow">
              <button className="primary" onClick={testRule}>测试规则</button>
              <span>{ruleResult?.riskLevel ?? 'Ready'}</span>
            </div>
            {ruleResult && (
              <div className="ruleResult">
                <div>
                  <b className={statusClass(ruleResult.riskLevel)}>{ruleResult.ruleCode ?? 'ALLOW'}</b>
                  <span>{ruleResult.action}</span>
                </div>
                <p>{ruleResult.reason}</p>
                <strong>{ruleResult.recommended}</strong>
              </div>
            )}
            <div className="ruleGrid">
              {rules.map((item) => (
                <article className="rule" key={item.id}>
                  <div>
                    <CheckCircle2 size={16} />
                    <strong>{item.name}</strong>
                  </div>
                  <span>{item.code}</span>
                  <p>{item.trigger}</p>
                </article>
              ))}
            </div>
          </section>
        </section>
      </section>
    </main>
  );
}

function statusClass(value?: string) {
  if (value === 'HIGH' || value === 'HIGH_PRIORITY' || value === 'CRITICAL') {
    return 'status danger';
  }
  if (value === 'MEDIUM' || value === 'KnowledgeGap') {
    return 'status warning';
  }
  return 'status';
}

createRoot(document.getElementById('root')!).render(<App />);
