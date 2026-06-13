import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  AlertTriangle,
  BookOpen,
  Bot,
  CheckCircle2,
  ClipboardCheck,
  Database,
  Send,
  FileSearch,
  FilePlus2,
  MessageSquareText,
  RefreshCcw,
  ShieldCheck,
  Download,
  UserRoundCheck
} from 'lucide-react';
import channelProtocolMatrix from '../../../contracts/channel-protocol-matrix.json';
import channelProtocolExamples from '../../../contracts/examples/channel-protocols.json';
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
  channel: string;
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
  annotationCount: number;
  averageReview: number;
  notes: string[];
};
type AnnotationDimensions = {
  groundedness: number;
  safety: number;
  helpfulness: number;
};
type Annotation = {
  id: string;
  messageId: string;
  reviewer: string;
  verdict: string;
  note: string;
  dimensions: AnnotationDimensions;
  tags: string[];
  score: number;
  createdAt: string;
};
type TrainingSample = {
  id: string;
  conversationId: string;
  messageId: string;
  channel: string;
  prompt: string;
  answer: string;
  engine: string;
  evidenceIds: string[];
  reviewer: string;
  verdict: string;
  score: number;
  dimensions: AnnotationDimensions;
  note: string;
  tags: string[];
  createdAt: string;
};
type ChannelPolicy = {
  channel: string;
  displayName: string;
  tone: string;
  slaMinutes: number;
  riskBoost: string;
  escalationNote: string;
  enabled: boolean;
};
type ChannelIntegration = {
  channel: string;
  displayName: string;
  enabled: boolean;
  secretSource: string;
  secretRef: string;
  nextSecretRef: string;
  signatureWindowSeconds: number;
  replayProtection: boolean;
  allowedOrigins: string[];
  rateLimitPerMinute: number;
  rotationHint: string;
  rotatesAt?: string;
  updatedAt: string;
};
type ChannelProtocolExample = {
  id: string;
  channel: string;
  endpoint: string;
  headers: Record<string, string>;
  secretRef: string;
  demoSecret: string;
  signatureInput: {
    channel: string;
    externalConversationId: string;
    timestamp: string;
    content: string;
  };
  request: Record<string, string>;
  expectedSuccess: {
    status: number;
    envelope: string;
  };
};
type ChannelProtocolMatrixRow = {
  channel: string;
  adapterEndpoint: string;
  successExampleId: string;
  conversationKey: string;
  messageKey: string;
  customerField: string;
  contentField: string;
  timestampField: string;
  origin: string;
  secretRef: string;
  replayKey: string;
  rateLimit: string;
  errors: string[];
};
type ChannelDemoResult = {
  exampleId: string;
  status: number;
  conversationId: string;
  reply: string;
  evidenceTitles: string[];
  fallbackReason?: string;
  trace?: AgentTrace;
};
type Dashboard = {
  metrics: Metric[];
  conversations: Conversation[] | null;
  knowledgeGaps: KnowledgeGap[] | null;
  rules: Rule[] | null;
  transfers: TransferTicket[] | null;
  channelPolicies: ChannelPolicy[] | null;
  integrations: ChannelIntegration[] | null;
  quality: QualitySummary;
  annotations: Annotation[] | null;
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

const hmacSHA256Hex = async (secret: string, payload: string) => {
  const encoder = new TextEncoder();
  const key = await crypto.subtle.importKey('raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
  const signature = await crypto.subtle.sign('HMAC', key, encoder.encode(payload));
  return Array.from(new Uint8Array(signature))
    .map((value) => value.toString(16).padStart(2, '0'))
    .join('');
};

const endpointPath = (endpoint: string) => endpoint.split(/\s+/)[1] ?? endpoint;

const cloneRecord = (value: Record<string, string>) => JSON.parse(JSON.stringify(value)) as Record<string, string>;

const signedDemoRequest = async (example: ChannelProtocolExample) => {
  const timestamp = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
  const unique = `${Date.now()}`;
  const request = cloneRecord(example.request);
  const signatureInput = { ...example.signatureInput, timestamp };

  if (example.endpoint.endsWith('/api/channels/inbound')) {
    request.externalConversationId = `${request.externalConversationId}-${unique}`;
    request.externalMessageId = `${request.externalMessageId}-${unique}`;
    request.timestamp = timestamp;
    signatureInput.externalConversationId = request.externalConversationId;
  } else if (example.endpoint.endsWith('/api/channels/wechat/inbound')) {
    request.openId = `${request.openId}-${unique}`;
    request.msgId = `${request.msgId}-${unique}`;
    request.timestamp = timestamp;
    signatureInput.externalConversationId = request.openId;
  } else if (example.endpoint.endsWith('/api/channels/app/inbound')) {
    request.deviceId = `${request.deviceId}-${unique}`;
    request.messageId = `${request.messageId}-${unique}`;
    request.sentAt = timestamp;
    signatureInput.externalConversationId = request.deviceId;
  } else if (example.endpoint.endsWith('/api/channels/marketplace/inbound')) {
    request.buyerId = `${request.buyerId}-${unique}`;
    request.eventId = `${request.eventId}-${unique}`;
    request.occurredAt = timestamp;
    signatureInput.externalConversationId = request.buyerId;
  }

  const payload = [
    signatureInput.channel,
    signatureInput.externalConversationId,
    signatureInput.timestamp,
    signatureInput.content
  ].map((item) => item.trim()).join('\n');
  request.signature = await hmacSHA256Hex(example.demoSecret, payload);
  return request;
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
  const [channelFilter, setChannelFilter] = useState('ALL');
  const [annotationVerdict, setAnnotationVerdict] = useState<'PASS' | 'REVIEW' | 'FAIL'>('PASS');
  const [annotationNote, setAnnotationNote] = useState('证据充分，回复安全，可作为教学正样本。');
  const [annotationSaving, setAnnotationSaving] = useState(false);
  const [trainingSamples, setTrainingSamples] = useState<TrainingSample[]>([]);
  const [channelDemoSending, setChannelDemoSending] = useState('');
  const [channelDemoResult, setChannelDemoResult] = useState<ChannelDemoResult | null>(null);

  const conversations = dashboard?.conversations ?? [];
  const gaps = dashboard?.knowledgeGaps ?? [];
  const rules = dashboard?.rules ?? [];
  const transfers = dashboard?.transfers ?? [];
  const annotations = dashboard?.annotations ?? [];
  const channelPolicies = dashboard?.channelPolicies ?? [];
  const integrations = dashboard?.integrations ?? [];
  const protocolExamples = channelProtocolExamples.examples as unknown as ChannelProtocolExample[];
  const protocolMatrix = channelProtocolMatrix.rows as ChannelProtocolMatrixRow[];
  const visibleConversations = conversations.filter((item) => channelFilter === 'ALL' || item.channel === channelFilter);
  const visibleTransfers = transfers.filter((ticket) => {
    if (channelFilter !== 'ALL' && ticket.channel !== channelFilter) {
      return false;
    }
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
    () => conversations.find((item) => item.id === selectedConversationId) ?? visibleConversations[0] ?? conversations[0],
    [conversations, selectedConversationId, visibleConversations]
  );
  const highRiskCount = conversations.filter((item) => item.riskLevel === 'HIGH').length;
  const openGapCount = gaps.filter((item) => item.status === 'OPEN').length;
  const openTransferCount = transfers.filter((item) => item.status === 'OPEN').length;
  const latestAssistantMessage = result?.agentMessage ?? [...history].reverse().find((item) => item.role === 'assistant');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [dashboardData, knowledgeData, sampleData] = await Promise.all([
        api<Dashboard>('/api/ops/dashboard'),
        api<KnowledgeArticle[]>('/api/knowledge/articles'),
        api<TrainingSample[]>('/api/ops/training-samples/export?maxScore=80')
      ]);
      setDashboard(dashboardData);
      setKnowledge(knowledgeData);
      setTrainingSamples(sampleData);
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

  const submitAnnotation = async () => {
    if (!latestAssistantMessage) {
      setError('请先选择或生成一条 Agent 回复');
      return;
    }
    setError('');
    setAnnotationSaving(true);
    try {
      await api<Annotation>('/api/ops/annotations/submit', {
        method: 'POST',
        body: JSON.stringify({
          messageId: latestAssistantMessage.id,
          reviewer: 'operator',
          verdict: annotationVerdict,
          note: annotationNote,
          dimensions: annotationDimensions(annotationVerdict),
          tags: ['human_review', 'teaching_sample']
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'submit annotation failed');
    } finally {
      setAnnotationSaving(false);
    }
  };

  const sendChannelDemo = async (example: ChannelProtocolExample) => {
    setError('');
    setChannelDemoSending(example.id);
    setChannelDemoResult(null);
    try {
      const request = await signedDemoRequest(example);
      const response = await fetch(endpointPath(example.endpoint), {
        method: 'POST',
        headers: example.headers,
        body: JSON.stringify(request)
      });
      const payload = (await response.json()) as ApiResponse<SendMessageResult>;
      if (!payload.success || !payload.data) {
        throw new Error(payload.error?.message ?? `channel demo failed: ${response.status}`);
      }
      setResult(payload.data);
      setSelectedConversationId(payload.data.conversation.id);
      setChannelFilter(example.channel);
      setChannelDemoResult({
        exampleId: example.id,
        status: response.status,
        conversationId: payload.data.conversation.id,
        reply: payload.data.agentMessage.content,
        evidenceTitles: payload.data.evidence.map((item) => item.title),
        fallbackReason: payload.data.agentMessage.fallbackReason,
        trace: payload.data.agentMessage.trace
      });
      await load();
      await loadMessages(payload.data.conversation.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'channel demo failed');
    } finally {
      setChannelDemoSending('');
    }
  };

  const downloadTrainingSamples = () => {
    const blob = new Blob([JSON.stringify(trainingSamples, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = 'agent-customer-service-training-samples.json';
    anchor.click();
    URL.revokeObjectURL(url);
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
              <span>{dashboard.quality.annotationCount} review</span>
              <span>{dashboard.quality.averageReview} avg</span>
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
            <div className="annotationBox">
              <div className="annotationHeader">
                <div>
                  <p className="sectionLabel">人工质检</p>
                  <strong>{latestAssistantMessage ? latestAssistantMessage.engine : '等待回复'}</strong>
                </div>
                <ClipboardCheck size={18} />
              </div>
              <div className="filterRow">
                {(['PASS', 'REVIEW', 'FAIL'] as const).map((item) => (
                  <button
                    className={annotationVerdict === item ? 'filterButton active' : 'filterButton'}
                    key={item}
                    onClick={() => setAnnotationVerdict(item)}
                  >
                    {item}
                  </button>
                ))}
              </div>
              <textarea
                className="annotationInput"
                value={annotationNote}
                onChange={(event) => setAnnotationNote(event.target.value)}
              />
              <div className="actionRow">
                <button className="primary" onClick={submitAnnotation} disabled={annotationSaving || !latestAssistantMessage}>
                  {annotationSaving ? '提交中' : '提交标注'}
                </button>
                <span>{dashboard?.quality.annotationCount ?? 0} reviews</span>
              </div>
            </div>
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
            <div className="filterRow">
              {['ALL', ...channelPolicies.map((item) => item.channel)].map((item) => (
                <button
                  className={channelFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => setChannelFilter(item)}
                >
                  {item}
                </button>
              ))}
            </div>
            <div className="tableList">
              {visibleConversations.map((item) => (
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
                    <em>{item.channel}</em>
                    <em>{item.intent}</em>
                    <b className={statusClass(item.riskLevel)}>{item.status}</b>
                  </div>
                </article>
              ))}
              {visibleConversations.length === 0 && <p className="empty">暂无匹配会话</p>}
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
                    <span>{ticket.channel} · {ticket.reason} · {ticket.conversationId}</span>
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
                <p className="sectionLabel">渠道接入</p>
                <h2>密钥与防重放</h2>
              </div>
              <span className="status">{integrations.length}</span>
            </div>
            <div className="tableList">
              {integrations.map((item) => (
                <article className="tableRow" key={item.channel}>
                  <div>
                    <strong>{item.displayName}</strong>
                    <span>{item.secretSource} · {item.secretRef}</span>
                    {item.nextSecretRef && <span>next · {item.nextSecretRef}</span>}
                    {item.allowedOrigins.length > 0 && <span>origins · {item.allowedOrigins.length}</span>}
                    <span>limit · {item.rateLimitPerMinute}/min</span>
                    <span>{item.rotationHint}</span>
                  </div>
                  <div>
                    <em>{item.rotatesAt ? item.rotatesAt.slice(0, 10) : `${item.signatureWindowSeconds}s`}</em>
                    <b className={statusClass(item.replayProtection ? 'PASS' : 'FAIL')}>{item.replayProtection ? 'REPLAY' : 'OFF'}</b>
                  </div>
                </article>
              ))}
              {integrations.length === 0 && <p className="empty">暂无渠道接入配置</p>}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">协议样例</p>
                <h2>真实渠道入口</h2>
              </div>
              <span className="status">{protocolExamples.length}</span>
            </div>
            <div className="protocolList">
              {protocolExamples.map((example) => (
                <article className="protocolExample" key={example.id}>
                  <div className="protocolSummary">
                    <strong>{example.channel}</strong>
                    <span>{example.endpoint}</span>
                    <span>{example.headers['X-Channel-Origin']}</span>
                    {channelDemoResult?.exampleId === example.id && (
                      <div className="protocolTrace">
                        <small>{channelDemoResult.reply}</small>
                        <span>
                          {[
                            channelDemoResult.trace?.strategy,
                            `${channelDemoResult.trace?.evidenceCount ?? channelDemoResult.evidenceTitles.length} evidence`,
                            `${channelDemoResult.trace?.historyCount ?? 0} history`,
                            channelDemoResult.fallbackReason ? `fallback ${channelDemoResult.fallbackReason}` : ''
                          ].filter(Boolean).join(' · ')}
                        </span>
                        {channelDemoResult.evidenceTitles.length > 0 && <span>{channelDemoResult.evidenceTitles.join(' / ')}</span>}
                      </div>
                    )}
                  </div>
                  <div className="protocolMeta">
                    <span>{example.secretRef}</span>
                    <span>{example.signatureInput.externalConversationId}</span>
                    <div className="protocolActions">
                      <button
                        className="tinyButton"
                        onClick={() => sendChannelDemo(example)}
                        title="发送演示请求"
                        disabled={channelDemoSending === example.id}
                      >
                        <Send size={14} />
                      </button>
                      <b className="status">{channelDemoResult?.exampleId === example.id ? channelDemoResult.status : example.expectedSuccess.status}</b>
                    </div>
                  </div>
                </article>
              ))}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">协议差异</p>
                <h2>字段映射与幂等</h2>
              </div>
              <span className="status">{protocolMatrix.length}</span>
            </div>
            <div className="protocolMatrix">
              {protocolMatrix.map((row) => (
                <article className="matrixRow" key={row.channel}>
                  <div>
                    <strong>{row.channel}</strong>
                    <span>{row.adapterEndpoint}</span>
                    <span>{row.conversationKey} / {row.messageKey} / {row.timestampField}</span>
                  </div>
                  <div>
                    <span>{row.origin}</span>
                    <span>{row.replayKey}</span>
                    <span>{row.rateLimit}</span>
                  </div>
                  <div>
                    <span>{row.contentField} to content</span>
                    <span>{row.customerField} to customer</span>
                    <span>{row.errors.slice(0, 3).join(' / ')}</span>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section className="panel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">渠道策略</p>
                <h2>多渠道 SLA</h2>
              </div>
              <span className="status">{channelPolicies.length}</span>
            </div>
            <div className="tableList">
              {channelPolicies.map((policy) => (
                <article className="tableRow" key={policy.channel}>
                  <div>
                    <strong>{policy.displayName}</strong>
                    <span>{policy.tone}</span>
                    <span>{policy.escalationNote}</span>
                  </div>
                  <div>
                    <em>{policy.slaMinutes}m</em>
                    <b className={statusClass(policy.riskBoost)}>{policy.riskBoost}</b>
                  </div>
                </article>
              ))}
              {channelPolicies.length === 0 && <p className="empty">暂无渠道策略</p>}
            </div>
          </section>

          <section className="panel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">质检审核</p>
                <h2>人工标注记录</h2>
              </div>
              <span className="status">{annotations.length}</span>
            </div>
            <div className="tableList">
              {annotations.slice(0, 5).map((item) => (
                <article className="tableRow" key={item.id}>
                  <div>
                    <strong>{item.verdict} · {item.score}</strong>
                    <span>{item.note}</span>
                    <span>{item.messageId}</span>
                  </div>
                  <div>
                    <em>{item.reviewer}</em>
                    <b className={statusClass(item.verdict)}>{item.dimensions.groundedness}/{item.dimensions.safety}/{item.dimensions.helpfulness}</b>
                  </div>
                </article>
              ))}
              {annotations.length === 0 && <p className="empty">暂无人工标注</p>}
            </div>
          </section>

          <section className="panel">
            <div className="panelHeader">
              <div>
                <p className="sectionLabel">复盘样本</p>
                <h2>低分导出</h2>
              </div>
              <button className="tinyButton" onClick={downloadTrainingSamples} title="导出样本" disabled={trainingSamples.length === 0}>
                <Download size={14} />
              </button>
            </div>
            <div className="tableList">
              {trainingSamples.slice(0, 5).map((item) => (
                <article className="tableRow" key={item.id}>
                  <div>
                    <strong>{item.verdict} · {item.score} · {item.channel}</strong>
                    <span>{item.prompt}</span>
                    <span>{item.note}</span>
                  </div>
                  <div>
                    <em>{item.engine}</em>
                    <b className={statusClass(item.verdict)}>{item.dimensions.groundedness}/{item.dimensions.safety}/{item.dimensions.helpfulness}</b>
                  </div>
                </article>
              ))}
              {trainingSamples.length === 0 && <p className="empty">暂无低分复盘样本</p>}
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
  if (value === 'HIGH' || value === 'HIGH_PRIORITY' || value === 'CRITICAL' || value === 'FAIL') {
    return 'status danger';
  }
  if (value === 'MEDIUM' || value === 'KnowledgeGap' || value === 'REVIEW') {
    return 'status warning';
  }
  return 'status';
}

function annotationDimensions(verdict: 'PASS' | 'REVIEW' | 'FAIL'): AnnotationDimensions {
  if (verdict === 'FAIL') {
    return { groundedness: 2, safety: 2, helpfulness: 2 };
  }
  if (verdict === 'REVIEW') {
    return { groundedness: 3, safety: 4, helpfulness: 3 };
  }
  return { groundedness: 5, safety: 5, helpfulness: 4 };
}

createRoot(document.getElementById('root')!).render(<App />);
