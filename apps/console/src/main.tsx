import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { AlertTriangle, BookOpen, Bot, MessagesSquare, RefreshCcw, ShieldCheck, UserRoundCheck } from 'lucide-react';
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
type Dashboard = {
  metrics: Metric[];
  conversations: Conversation[];
  knowledgeGaps: KnowledgeGap[];
  rules: Rule[];
};
type SendMessageResult = {
  conversation: Conversation;
  agentMessage: { content: string; engine: string; fallbackReason?: string; evidenceIds?: string[] };
  evidence: { id: string; title: string; category: string; trustLevel: string }[];
  gap?: KnowledgeGap;
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

function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState('7 天无理由退货的运费怎么计算？');
  const [result, setResult] = useState<SendMessageResult | null>(null);
  const [error, setError] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      setDashboard(await api<Dashboard>('/api/ops/dashboard'));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'load failed');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const activeConversation = useMemo(() => dashboard?.conversations[0], [dashboard]);

  const send = async () => {
    setError('');
    try {
      const data = await api<SendMessageResult>('/api/customer-service/messages', {
        method: 'POST',
        body: JSON.stringify({ conversationId: activeConversation?.id, content: message })
      });
      setResult(data);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'send failed');
    }
  };

  return (
    <main className="shell">
      <aside className="nav">
        <div className="brand">
          <Bot size={24} />
          <span>Agent CS</span>
        </div>
        <a className="active"><MessagesSquare size={18} /> 会话</a>
        <a><BookOpen size={18} /> 知识</a>
        <a><ShieldCheck size={18} /> 规则</a>
        <a><UserRoundCheck size={18} /> 人工</a>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">DVSkyFolding Runtime</p>
            <h1>可靠客服 Agent 控制台</h1>
          </div>
          <button className="iconButton" onClick={load} aria-label="刷新" title="刷新">
            <RefreshCcw size={18} />
          </button>
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
          {loading && <article className="metric skeleton">Loading</article>}
        </section>

        <section className="grid">
          <div className="panel conversationPanel">
            <div className="panelHeader">
              <h2>Agent 工作区</h2>
              <span className="status">{activeConversation?.status ?? 'Ready'}</span>
            </div>
            <textarea value={message} onChange={(event) => setMessage(event.target.value)} />
            <button className="primary" onClick={send}>发送并审计</button>
            {result && (
              <div className="reply">
                <div className="replyMeta">
                  <span>{result.agentMessage.engine}</span>
                  <span>{result.agentMessage.fallbackReason ?? 'EVIDENCE_OK'}</span>
                </div>
                <p>{result.agentMessage.content}</p>
              </div>
            )}
          </div>

          <div className="panel">
            <div className="panelHeader"><h2>会话队列</h2></div>
            <div className="list">
              {(dashboard?.conversations ?? []).map((item) => (
                <article className="row" key={item.id}>
                  <div>
                    <strong>{item.customer}</strong>
                    <span>{item.intent} · {item.channel}</span>
                  </div>
                  <em className={item.riskLevel === 'HIGH' ? 'danger' : ''}>{item.status}</em>
                </article>
              ))}
            </div>
          </div>

          <div className="panel">
            <div className="panelHeader"><h2>知识缺口</h2></div>
            <div className="list">
              {(dashboard?.knowledgeGaps ?? []).map((item) => (
                <article className="row" key={item.id}>
                  <div>
                    <strong>{item.question}</strong>
                    <span>{item.reason}</span>
                  </div>
                  <em>{item.priority}</em>
                </article>
              ))}
              {dashboard?.knowledgeGaps.length === 0 && <p className="empty">暂无开放缺口</p>}
            </div>
          </div>

          <div className="panel">
            <div className="panelHeader"><h2>兜底规则</h2></div>
            <div className="list">
              {(dashboard?.rules ?? []).map((item) => (
                <article className="row" key={item.id}>
                  <div>
                    <strong>{item.name}</strong>
                    <span>{item.code}</span>
                  </div>
                  <em>{item.enabled ? 'ON' : 'OFF'}</em>
                </article>
              ))}
            </div>
          </div>
        </section>
      </section>
    </main>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
