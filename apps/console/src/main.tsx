import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  AlertTriangle,
  BookOpen,
  Bot,
  CheckCircle2,
  CircleX,
  ClipboardCheck,
  Database,
  Send,
  FileSearch,
  FilePlus2,
  MessageSquareText,
  RefreshCcw,
  Save,
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
type Rule = {
  id: string;
  code: string;
  name: string;
  trigger: string;
  action: string;
  enabled: boolean;
  version: string;
  stage: string;
  hitCount: number;
  lastHitAt?: string;
};
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
type RuleComparison = {
  input: string;
  current: RuleTestResult;
  canary: RuleTestResult;
  changed: boolean;
  recommendation: string;
};
type RuleReleaseEvent = {
  id: string;
  ruleCode: string;
  version: string;
  action: string;
  actor: string;
  note: string;
  createdAt: string;
};
type NotificationPolicyEvent = {
  id: string;
  channel: string;
  action: string;
  actor: string;
  before: string;
  after: string;
  note: string;
  createdAt: string;
};
type NotificationPolicyDiff = {
  field: string;
  before: string;
  after: string;
};
type NotificationPolicyChange = {
  id: string;
  channel: string;
  targetUrl: string;
  secretRef: string;
  maxAttempts: number;
  backoffSeconds: number;
  currentTargetUrl: string;
  currentSecretRef: string;
  currentMaxAttempts: number;
  currentBackoffSeconds: number;
  diff: NotificationPolicyDiff[];
  confirmationText: string;
  requestedBy: string;
  status: string;
  note: string;
  createdAt: string;
  expiresAt: string;
  approvedBy?: string;
  approvedAt?: string;
};
type RuleReleaseObservation = {
  ruleCode: string;
  version: string;
  releasedAt: string;
  window: string;
  ruleHits: number;
  transferTickets: number;
  lowScoreSamples: number;
  averageReview: number;
  riskLevel: string;
  recommendation: string;
  rollbackRecommended: boolean;
};
type RuleApproval = {
  id: string;
  ruleCode: string;
  approver: string;
  riskLevel: string;
  sampleIds: string[];
  sampleCount: number;
  status: string;
  note: string;
  createdAt: string;
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
type ReviewTask = {
  id: string;
  messageId: string;
  conversationId: string;
  channel: string;
  assignee?: string;
  status: string;
  priority: string;
  reason: string;
  createdAt: string;
  completedAt?: string;
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
type ChannelAlert = {
  channel: string;
  code: string;
  count: number;
  lastReason: string;
  lastOrigin: string;
  lastSeenAt: string;
};
type ChannelFailureTrend = {
  channel: string;
  bucketStart: string;
  count: number;
};
type ChannelInboundAudit = {
  id: string;
  channel: string;
  externalConversationId: string;
  externalMessageId?: string;
  origin?: string;
  status: string;
  code: string;
  reason?: string;
  replayKey?: string;
  signaturePreview?: string;
  contentHash?: string;
  createdAt: string;
};
type ChannelAlertPolicy = {
  channel: string;
  severity: string;
  threshold: number;
  windowMinutes: number;
  notifyTarget: string;
  targetUrl: string;
  secretRef: string;
  maxAttempts: number;
  backoffSeconds: number;
  inboundAuditMinSamples: number;
  inboundAuditMinAcceptanceRate: number;
  inboundAuditMaxErrorCount: number;
  enabled: boolean;
  active: boolean;
  currentCount: number;
  lastTriggeredAt?: string;
};
type NotificationPolicyDraft = {
  targetUrl: string;
  secretRef: string;
  maxAttempts: number;
  backoffSeconds: number;
  inboundAuditMinSamples: number;
  inboundAuditMinAcceptanceRate: number;
  inboundAuditMaxErrorCount: number;
};
type ChannelNotification = {
  id: string;
  channel: string;
  severity: string;
  target: string;
  targetUrl: string;
  secretRef: string;
  status: string;
  reason: string;
  count: number;
  attempts: number;
  maxAttempts: number;
  backoffSeconds: number;
  nextRetryAt?: string;
  signature?: string;
  lastDispatchAt?: string;
  lastError?: string;
  receiptStatus?: string;
  receiptBody?: string;
  deliveryAudit?: {
    attempt: number;
    targetUrl: string;
    secretRef: string;
    signaturePreview: string;
    payloadHash: string;
    requestSummary: string;
    responseSummary: string;
    createdAt: string;
  }[];
  deadLetterReason?: string;
  createdAt: string;
  ackedBy?: string;
  ackNote?: string;
  ackedAt?: string;
};
type ChannelRunbookCheck = {
  id: string;
  channel: string;
  runbookStatus: string;
  checkStatus: string;
  step: string;
  stepIndex: number;
  actionRef?: string;
  reportId?: string;
  assignee?: string;
  dueAt?: string;
  actor: string;
  note?: string;
  completedAt: string;
};

type ChannelRunbook = {
  channel: string;
  severity: string;
  status: string;
  failureCode: string;
  owner: string;
  nextAction: string;
  escalation: string;
  notificationId?: string;
  notificationState?: string;
  steps: string[];
  checks?: ChannelRunbookCheck[] | null;
  checkSummary: {
    total: number;
    done: number;
    blocked: number;
    todo: number;
  };
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
type ChannelSignatureProfile = {
  id: string;
  channel: string;
  adapterEndpoint: string;
  signatureHeader: string;
  timestampHeader: string;
  replayHeader: string;
  canonicalPayload: string[];
  sampleSignature: string;
  retrySemantics: string;
  failureCodes: string[];
};
type ChannelProtocolMatrixRow = {
  channel: string;
  adapterEndpoint: string;
  successExampleId: string;
  signatureProfileId: string;
  conversationKey: string;
  messageKey: string;
  customerField: string;
  contentField: string;
  timestampField: string;
  origin: string;
  secretRef: string;
  signatureHeader: string;
  timestampHeader: string;
  replayKey: string;
  rateLimit: string;
  retrySemantics: string;
  errors: string[];
};
type ChannelErrorExample = {
  id: string;
  exampleId: string;
  status: number;
  code: string;
  mutation: 'origin' | 'signature' | 'timestamp' | 'volume' | 'duplicate';
  reason: string;
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
type ChannelFailureResult = {
  exampleId: string;
  status: number;
  code: string;
  reason: string;
};
type ChannelOpsHandoffPriority = {
  rank: number;
  channel: string;
  severity: string;
  source: string;
  reason: string;
  recommendedAction: string;
  count: number;
  actionType: string;
  actionRef?: string;
  actionLabel?: string;
  notificationId?: string;
  runbookStatus?: string;
};

type ChannelOpsReportSummary = {
  failureCount: number;
  activeRunbooks: number;
  openNotifications: number;
  retrying: number;
  deadLetters: number;
  channels: string[];
  handoffPriorities?: ChannelOpsHandoffPriority[] | null;
  inboundAudit?: {
    total: number;
    accepted: number;
    rejected: number;
    acceptanceRate: number;
    topErrorCodes: Array<{ code: string; count: number }>;
  };
  inboundAuditQuality?: {
    eventCount: number;
    active: number;
    watch: number;
    recovered: number;
    activeChannels: string[] | null;
    watchChannels: string[] | null;
    recoveredChannels: string[] | null;
  };
};
type ChannelOpsReport = {
  id: string;
  format: 'markdown' | 'csv';
  contentType: string;
  content?: string;
  summary: ChannelOpsReportSummary;
  generatedAt: string;
};
type ChannelOpsReportScheduler = {
  enabled: boolean;
  format: 'markdown' | 'csv';
  intervalMins: number;
  retain: number;
  runOnStart: boolean;
  lastRunAt?: string;
  nextRunAt?: string;
  lastReportId?: string;
  lastStatus: 'DISABLED' | 'PENDING' | 'SUCCESS' | 'FAILED';
  lastError?: string;
  lastPruned: number;
};
type ChannelOpsReportEvent = {
  id: string;
  action: string;
  actor: string;
  status: 'SUCCESS' | 'FAILED';
  reportId?: string;
  format: 'markdown' | 'csv';
  pruned: number;
  note?: string;
  error?: string;
  createdAt: string;
};
type ChannelInboundAuditQualityEvent = {
  id: string;
  channel: string;
  severity: string;
  status: 'WATCH' | 'ESCALATE';
  failureCode: string;
  total: number;
  accepted: number;
  rejected: number;
  acceptanceRate: number;
  minSamples: number;
  minAcceptanceRate: number;
  maxErrorCount: number;
  reason: string;
  createdAt: string;
};
type ChannelOpsReportCompensationResult = {
  event: ChannelOpsReportEvent;
  report?: ChannelOpsReport;
  status: ChannelOpsReportScheduler;
};
type Dashboard = {
  metrics: Metric[];
  conversations: Conversation[] | null;
  knowledgeGaps: KnowledgeGap[] | null;
  rules: Rule[] | null;
  ruleApprovals: RuleApproval[] | null;
  ruleEvents: RuleReleaseEvent[] | null;
  ruleObservations: RuleReleaseObservation[] | null;
  transfers: TransferTicket[] | null;
  channelPolicies: ChannelPolicy[] | null;
  integrations: ChannelIntegration[] | null;
  channelAlerts: ChannelAlert[] | null;
  channelFailureTrends: ChannelFailureTrend[] | null;
  channelInboundAudits: ChannelInboundAudit[] | null;
  channelInboundAuditQualityEvents: ChannelInboundAuditQualityEvent[] | null;
  channelAlertPolicies: ChannelAlertPolicy[] | null;
  channelNotifications: ChannelNotification[] | null;
  channelRunbooks: ChannelRunbook[] | null;
  notificationPolicyEvents: NotificationPolicyEvent[] | null;
  notificationPolicyChanges: NotificationPolicyChange[] | null;
  quality: QualitySummary;
  annotations: Annotation[] | null;
  reviewTasks: ReviewTask[] | null;
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
  } else if (example.endpoint.endsWith('/api/channels/wecom/inbound')) {
    request.userId = `${request.userId}-${unique}`;
    request.msgId = `${request.msgId}-${unique}`;
    request.eventTime = timestamp;
    signatureInput.externalConversationId = `${request.corpId}:${request.userId}`;
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

const errorHeaders = (example: ChannelProtocolExample, error: ChannelErrorExample) => {
  if (error.mutation === 'origin') {
    return { ...example.headers, 'X-Channel-Origin': 'https://evil.example.com' };
  }
  return example.headers;
};

const staleTimestamp = () => new Date(Date.now() - 10 * 60 * 1000).toISOString().replace(/\.\d{3}Z$/, 'Z');

const failedDemoRequest = async (example: ChannelProtocolExample, error: ChannelErrorExample) => {
  const request = await signedDemoRequest(example);
  if (error.mutation === 'signature') {
    request.signature = 'bad-signature';
  }
  if (error.mutation === 'timestamp') {
    const stale = staleTimestamp();
    request.timestamp = stale;
    const payload = [request.channel, request.externalConversationId, stale, request.content].map((item) => item.trim()).join('\n');
    request.signature = await hmacSHA256Hex(example.demoSecret, payload);
  }
  return request;
};

function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [knowledge, setKnowledge] = useState<KnowledgeArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState('7 天无理由退货的运费怎么计算？');
  const [ruleInput, setRuleInput] = useState('我已经投诉很多次了，现在必须转人工');
  const [ruleResult, setRuleResult] = useState<RuleTestResult | null>(null);
  const [ruleComparison, setRuleComparison] = useState<RuleComparison | null>(null);
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
  const [channelOpsReports, setChannelOpsReports] = useState<ChannelOpsReport[]>([]);
  const [reportScheduler, setReportScheduler] = useState<ChannelOpsReportScheduler | null>(null);
  const [reportEvents, setReportEvents] = useState<ChannelOpsReportEvent[]>([]);
  const [auditQualityEvents, setAuditQualityEvents] = useState<ChannelInboundAuditQualityEvent[]>([]);
  const [runbookCheckRows, setRunbookCheckRows] = useState<ChannelRunbookCheck[]>([]);
  const [reportEventStatusFilter, setReportEventStatusFilter] = useState<'ALL' | 'SUCCESS' | 'FAILED'>('ALL');
  const [reportEventActorFilter, setReportEventActorFilter] = useState('');
  const [runbookCheckChannelFilter, setRunbookCheckChannelFilter] = useState('ALL');
  const [runbookCheckStatusFilter, setRunbookCheckStatusFilter] = useState('ALL');
  const [runbookCheckActorFilter, setRunbookCheckActorFilter] = useState('');
  const [reportGenerating, setReportGenerating] = useState('');
  const [reportCompensating, setReportCompensating] = useState(false);
  const [channelDemoSending, setChannelDemoSending] = useState('');
  const [channelDemoResult, setChannelDemoResult] = useState<ChannelDemoResult | null>(null);
  const [channelFailureSending, setChannelFailureSending] = useState('');
  const [channelFailureResult, setChannelFailureResult] = useState<ChannelFailureResult | null>(null);
  const [inboundAuditStatusFilter, setInboundAuditStatusFilter] = useState<'ALL' | 'ACCEPTED' | 'REJECTED'>('ALL');
  const [inboundAuditChannelFilter, setInboundAuditChannelFilter] = useState('ALL');
  const [inboundAuditCodeFilter, setInboundAuditCodeFilter] = useState('');
  const [auditQualityEventStatusFilter, setAuditQualityEventStatusFilter] = useState<'ALL' | 'WATCH' | 'ESCALATE'>('ALL');
  const [notificationStatusFilter, setNotificationStatusFilter] = useState<'ALL' | 'OPEN' | 'RETRYING' | 'SENT' | 'DEAD_LETTER' | 'ACKED'>('ALL');
  const [notificationChannelFilter, setNotificationChannelFilter] = useState('ALL');
  const [expandedNotificationId, setExpandedNotificationId] = useState('');
  const [notificationPolicyDrafts, setNotificationPolicyDrafts] = useState<Record<string, NotificationPolicyDraft>>({});

  const conversations = dashboard?.conversations ?? [];
  const gaps = dashboard?.knowledgeGaps ?? [];
  const rules = dashboard?.rules ?? [];
  const ruleApprovals = dashboard?.ruleApprovals ?? [];
  const ruleEvents = dashboard?.ruleEvents ?? [];
  const ruleObservations = dashboard?.ruleObservations ?? [];
  const transfers = dashboard?.transfers ?? [];
  const annotations = dashboard?.annotations ?? [];
  const reviewTasks = dashboard?.reviewTasks ?? [];
  const channelPolicies = dashboard?.channelPolicies ?? [];
  const integrations = dashboard?.integrations ?? [];
  const channelAlerts = dashboard?.channelAlerts ?? [];
  const channelFailureTrends = dashboard?.channelFailureTrends ?? [];
  const channelInboundAudits = dashboard?.channelInboundAudits ?? [];
  const channelAlertPolicies = dashboard?.channelAlertPolicies ?? [];
  const channelNotifications = dashboard?.channelNotifications ?? [];
  const channelRunbooks = dashboard?.channelRunbooks ?? [];
  const runbookCheckChannels = Array.from(new Set([...channelRunbooks.map((item) => item.channel), ...runbookCheckRows.map((item) => item.channel)])).sort();
  const runbookCheckStatuses = Array.from(new Set([...channelRunbooks.map((item) => item.status), ...runbookCheckRows.map((item) => item.runbookStatus)])).sort();
  const acknowledgedNotificationIds = new Set(channelNotifications.filter((item) => item.status === 'ACKED' || item.ackedBy).map((item) => item.id));
  const notificationPolicyEvents = dashboard?.notificationPolicyEvents ?? [];
  const notificationPolicyChanges = dashboard?.notificationPolicyChanges ?? [];
  const protocolExamples = channelProtocolExamples.examples as unknown as ChannelProtocolExample[];
  const signatureProfiles = channelProtocolExamples.platformSignatureProfiles as unknown as ChannelSignatureProfile[];
  const errorExamples = channelProtocolExamples.errorExamples as ChannelErrorExample[];
  const protocolMatrix = channelProtocolMatrix.rows as ChannelProtocolMatrixRow[];
  const auditStatusOptions = ['ALL', 'ACCEPTED', 'REJECTED'] as const;
  const auditQualityEventStatusOptions = ['ALL', 'WATCH', 'ESCALATE'] as const;
  const auditChannels = Array.from(new Set([...protocolMatrix.map((item) => item.channel), ...channelInboundAudits.map((item) => item.channel), ...auditQualityEvents.map((item) => item.channel)])).sort();
  const alertPolicyByChannel = Object.fromEntries(channelAlertPolicies.map((policy) => [policy.channel, policy]));
  const auditQualityRows = auditChannels.map((channel) => {
    const items = channelInboundAudits.filter((audit) => audit.channel === channel);
    const total = items.length;
    const accepted = items.filter((audit) => audit.status === 'ACCEPTED').length;
    const rejected = items.filter((audit) => audit.status === 'REJECTED').length;
    const acceptanceRate = total > 0 ? Math.round((accepted / total) * 100) : 0;
    const topCode = items
      .filter((audit) => audit.status === 'REJECTED')
      .reduce<Record<string, number>>((acc, audit) => {
        acc[audit.code] = (acc[audit.code] ?? 0) + 1;
        return acc;
      }, {});
    const topError = Object.entries(topCode).sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))[0];
    return { channel, total, accepted, rejected, acceptanceRate, topErrorCode: topError?.[0] ?? '', topErrorCount: topError?.[1] ?? 0, topError: topError ? `${topError[0]}:${topError[1]}` : '' };
  }).filter((item) => item.total > 0);
  const auditQualityByChannel = Object.fromEntries(auditQualityRows.map((row) => [row.channel, row]));
  const latestAuditEventByChannel = auditQualityEvents.reduce<Record<string, ChannelInboundAuditQualityEvent>>((acc, event) => {
    const current = acc[event.channel];
    if (!current || event.createdAt > current.createdAt) {
      acc[event.channel] = event;
    }
    return acc;
  }, {});
  const auditRecoveryRows = Object.values(latestAuditEventByChannel).map((event) => {
    const current = auditQualityByChannel[event.channel];
    const policy = alertPolicyByChannel[event.channel];
    const minSamples = policy?.inboundAuditMinSamples ?? event.minSamples;
    const minAcceptanceRate = policy?.inboundAuditMinAcceptanceRate ?? event.minAcceptanceRate;
    const maxErrorCount = policy?.inboundAuditMaxErrorCount ?? event.maxErrorCount;
    const currentRate = current?.acceptanceRate ?? event.acceptanceRate;
    const currentTotal = current?.total ?? 0;
    const currentTopErrorCount = current?.topErrorCount ?? 0;
    const recovered = currentTotal >= minSamples && currentRate >= minAcceptanceRate && currentTopErrorCount < maxErrorCount;
    const state = recovered ? 'RECOVERED' : event.status === 'ESCALATE' ? 'ACTIVE' : 'WATCH';
    return {
      channel: event.channel,
      state,
      failureCode: event.failureCode,
      eventRate: event.acceptanceRate,
      currentRate,
      currentTotal,
      minSamples,
      minAcceptanceRate,
      maxErrorCount,
      createdAt: event.createdAt
    };
  }).sort((left, right) => {
    if (left.state !== right.state) {
      return left.state === 'ACTIVE' ? -1 : right.state === 'ACTIVE' ? 1 : left.state === 'WATCH' ? -1 : 1;
    }
    return right.createdAt.localeCompare(left.createdAt);
  });
  const notificationChannels = Array.from(new Set(channelNotifications.map((item) => item.channel))).sort();
  const notificationStatusOptions = ['ALL', 'OPEN', 'RETRYING', 'SENT', 'DEAD_LETTER', 'ACKED'] as const;
  const visibleNotifications = channelNotifications.filter((item) => {
    if (notificationStatusFilter !== 'ALL' && item.status !== notificationStatusFilter) {
      return false;
    }
    if (notificationChannelFilter !== 'ALL' && item.channel !== notificationChannelFilter) {
      return false;
    }
    return true;
  });
  const notificationStats = {
    open: channelNotifications.filter((item) => item.status === 'OPEN').length,
    retrying: channelNotifications.filter((item) => item.status === 'RETRYING').length,
    dead: channelNotifications.filter((item) => item.status === 'DEAD_LETTER').length,
    audited: channelNotifications.filter((item) => (item.deliveryAudit?.length ?? 0) > 0).length
  };
  const maxTrendCount = Math.max(1, ...channelFailureTrends.map((item) => item.count));
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
  const trainingSampleById = useMemo(() => new Map(trainingSamples.map((item) => [item.id, item])), [trainingSamples]);
  const approvalSampleCandidates = trainingSamples.slice(0, 3);
  const highRiskCount = conversations.filter((item) => item.riskLevel === 'HIGH').length;
  const openGapCount = gaps.filter((item) => item.status === 'OPEN').length;
  const openTransferCount = transfers.filter((item) => item.status === 'OPEN').length;
  const openReviewCount = reviewTasks.filter((item) => item.status !== 'COMPLETED').length;
  const latestAssistantMessage = result?.agentMessage ?? [...history].reverse().find((item) => item.role === 'assistant');
  const hasApprovedRuleGate = (rule: Rule) => ruleApprovals.some((item) => item.ruleCode === rule.code && item.status === 'APPROVED' && item.sampleCount >= 3);
  const reportEventStatusOptions = ['ALL', 'SUCCESS', 'FAILED'] as const;

  const channelOpsEventQuery = (status = reportEventStatusFilter, actor = reportEventActorFilter, limit = 4) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (status !== 'ALL') {
      params.set('status', status);
    }
    if (actor.trim()) {
      params.set('actor', actor.trim());
    }
    return params.toString();
  };

  const refreshReportEvents = async (status = reportEventStatusFilter, actor = reportEventActorFilter) => {
    const events = await api<ChannelOpsReportEvent[]>(`/api/ops/channel-ops-report-events?${channelOpsEventQuery(status, actor)}`);
    setReportEvents(events);
  };

  const channelRunbookCheckQuery = (
    channel = runbookCheckChannelFilter,
    status = runbookCheckStatusFilter,
    actor = runbookCheckActorFilter,
    limit = 20
  ) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (channel !== 'ALL') {
      params.set('channel', channel);
    }
    if (status !== 'ALL') {
      params.set('status', status);
    }
    if (actor.trim()) {
      params.set('actor', actor.trim());
    }
    return params.toString();
  };

  const refreshRunbookChecks = async (
    channel = runbookCheckChannelFilter,
    status = runbookCheckStatusFilter,
    actor = runbookCheckActorFilter
  ) => {
    const checks = await api<ChannelRunbookCheck[]>(`/api/ops/channel-runbook-checks?${channelRunbookCheckQuery(channel, status, actor)}`);
    setRunbookCheckRows(checks);
  };

  const channelInboundAuditQuery = (
    channel = inboundAuditChannelFilter,
    status = inboundAuditStatusFilter,
    code = inboundAuditCodeFilter,
    limit = 20
  ) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (channel !== 'ALL') {
      params.set('channel', channel);
    }
    if (status !== 'ALL') {
      params.set('status', status);
    }
    if (code.trim()) {
      params.set('code', code.trim());
    }
    return params.toString();
  };

  const refreshInboundAudits = async (
    channel = inboundAuditChannelFilter,
    status = inboundAuditStatusFilter,
    code = inboundAuditCodeFilter
  ) => {
    const audits = await api<ChannelInboundAudit[]>(`/api/ops/channel-inbound-audits?${channelInboundAuditQuery(channel, status, code)}`);
    setDashboard((current) => current ? { ...current, channelInboundAudits: audits } : current);
  };

  const channelInboundAuditQualityEventQuery = (
    channel = inboundAuditChannelFilter,
    status: 'ALL' | 'WATCH' | 'ESCALATE' = 'ALL',
    code = inboundAuditCodeFilter,
    limit = 20
  ) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (channel !== 'ALL') {
      params.set('channel', channel);
    }
    if (status !== 'ALL') {
      params.set('status', status);
    }
    if (code.trim()) {
      params.set('code', code.trim());
    }
    return params.toString();
  };

  const refreshAuditQualityEvents = async (
    channel = inboundAuditChannelFilter,
    status: 'ALL' | 'WATCH' | 'ESCALATE' = 'ALL',
    code = inboundAuditCodeFilter
  ) => {
    const events = await api<ChannelInboundAuditQualityEvent[]>(`/api/ops/channel-inbound-audit-quality-events?${channelInboundAuditQualityEventQuery(channel, status, code)}`);
    setAuditQualityEvents(events);
  };

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [dashboardData, knowledgeData, sampleData, reportData, schedulerData, reportEventData, auditQualityEventData, runbookCheckData] = await Promise.all([
        api<Dashboard>('/api/ops/dashboard'),
        api<KnowledgeArticle[]>('/api/knowledge/articles'),
        api<TrainingSample[]>('/api/ops/training-samples/export?maxScore=80'),
        api<ChannelOpsReport[]>('/api/ops/channel-ops-reports?limit=6'),
        api<ChannelOpsReportScheduler>('/api/ops/channel-ops-report-scheduler'),
        api<ChannelOpsReportEvent[]>(`/api/ops/channel-ops-report-events?${channelOpsEventQuery()}`),
        api<ChannelInboundAuditQualityEvent[]>(`/api/ops/channel-inbound-audit-quality-events?${channelInboundAuditQualityEventQuery()}`),
        api<ChannelRunbookCheck[]>(`/api/ops/channel-runbook-checks?${channelRunbookCheckQuery()}`)
      ]);
      setDashboard(dashboardData);
      setKnowledge(knowledgeData ?? []);
      setTrainingSamples(sampleData ?? []);
      setChannelOpsReports(reportData ?? []);
      setReportScheduler(schedulerData);
      setReportEvents(reportEventData ?? []);
      setAuditQualityEvents(auditQualityEventData ?? []);
      setRunbookCheckRows(runbookCheckData ?? []);
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

  useEffect(() => {
    const nextDrafts = Object.fromEntries(channelAlertPolicies.map((policy) => [
      policy.channel,
      {
        targetUrl: policy.targetUrl,
        secretRef: policy.secretRef,
        maxAttempts: policy.maxAttempts,
        backoffSeconds: policy.backoffSeconds,
        inboundAuditMinSamples: policy.inboundAuditMinSamples,
        inboundAuditMinAcceptanceRate: policy.inboundAuditMinAcceptanceRate,
        inboundAuditMaxErrorCount: policy.inboundAuditMaxErrorCount
      }
    ]));
    setNotificationPolicyDrafts(nextDrafts);
  }, [dashboard?.channelAlertPolicies]);

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

  const compareRules = async () => {
    setError('');
    try {
      setRuleComparison(await api<RuleComparison>('/api/ops/rules/compare', {
        method: 'POST',
        body: JSON.stringify({ content: ruleInput })
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'compare rules failed');
    }
  };

  const approveRuleRelease = async (rule: Rule) => {
    setError('');
    const sampleIds = approvalSampleCandidates.map((item) => item.id);
    const fallbackSampleIds = ['sample_cancel_1', 'sample_cancel_2', 'sample_cancel_3'];
    const hasTrainingSamples = sampleIds.length >= 3;
    try {
      await api<RuleApproval>('/api/ops/rules/approve', {
        method: 'POST',
        body: JSON.stringify({
          code: rule.code,
          approver: 'qa-lead',
          riskLevel: 'LOW',
          sampleIds: hasTrainingSamples ? sampleIds : fallbackSampleIds,
          note: hasTrainingSamples ? '灰度对比通过，已绑定低分复盘样本' : '灰度对比通过，使用课堂样例样本 ID'
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'approve rule failed');
    }
  };

  const publishCanaryRule = async (rule: Rule) => {
    setError('');
    try {
      await api<RuleReleaseEvent>('/api/ops/rules/publish-canary', {
        method: 'POST',
        body: JSON.stringify({ code: rule.code, actor: 'qa-lead', note: '灰度对比通过，发布为 active 规则' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'publish rule failed');
    }
  };

  const rollbackRule = async (rule: Rule) => {
    setError('');
    try {
      await api<RuleReleaseEvent>('/api/ops/rules/rollback', {
        method: 'POST',
        body: JSON.stringify({ code: rule.code, actor: 'qa-lead', note: '发布后观测异常，回滚规则' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'rollback rule failed');
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

  const assignReviewTask = async (task: ReviewTask) => {
    setError('');
    try {
      await api<ReviewTask>('/api/ops/review-tasks/assign', {
        method: 'POST',
        body: JSON.stringify({ id: task.id, assignee: 'qa-operator' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'assign review task failed');
    }
  };

  const completeReviewTask = async (task: ReviewTask) => {
    setError('');
    try {
      await api<ReviewTask>('/api/ops/review-tasks/complete', {
        method: 'POST',
        body: JSON.stringify({ id: task.id })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'complete review task failed');
    }
  };

  const acknowledgeChannelNotification = async (notification: ChannelNotification) => {
    setError('');
    try {
      await api<ChannelNotification>('/api/ops/channel-notifications/ack', {
        method: 'POST',
        body: JSON.stringify({ id: notification.id, actor: 'ops-a', note: '已确认并通知渠道负责人' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'ack notification failed');
    }
  };

  const acknowledgeHandoffPriority = async (report: ChannelOpsReport, priority: ChannelOpsHandoffPriority) => {
    if (!priority.notificationId) {
      setError('handoff priority has no notification to acknowledge');
      return;
    }
    setError('');
    try {
      await api<ChannelNotification>('/api/ops/channel-notifications/ack', {
        method: 'POST',
        body: JSON.stringify({
          id: priority.notificationId,
          actor: 'ops-a',
          note: `handoff ${report.id} ${priority.actionRef ?? priority.source}: ${priority.actionLabel ?? priority.recommendedAction}`
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'ack handoff failed');
    }
  };

  const updateRunbookCheckStatus = async (
    runbook: ChannelRunbook,
    step: string,
    stepIndex: number,
    path: '/api/ops/channel-runbook-checks/complete' | '/api/ops/channel-runbook-checks/block' | '/api/ops/channel-runbook-checks/recover',
    note: string
  ) => {
    setError('');
    try {
      await api<ChannelRunbookCheck>(path, {
        method: 'POST',
        body: JSON.stringify({
          channel: runbook.channel,
          runbookStatus: runbook.status,
          step,
          stepIndex,
          actionRef: `${runbook.channel}:${runbook.status}`,
          assignee: runbook.owner,
          dueAt: new Date(Date.now() + 4 * 60 * 60 * 1000).toISOString(),
          actor: 'ops-a',
          note
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'update runbook check failed');
    }
  };

  const completeRunbookCheck = async (runbook: ChannelRunbook, step: string, stepIndex: number) => updateRunbookCheckStatus(
    runbook,
    step,
    stepIndex,
    '/api/ops/channel-runbook-checks/complete',
    `completed ${runbook.channel}:${runbook.status} step ${stepIndex + 1}`
  );

  const blockRunbookCheck = async (runbook: ChannelRunbook, step: string, stepIndex: number) => updateRunbookCheckStatus(
    runbook,
    step,
    stepIndex,
    '/api/ops/channel-runbook-checks/block',
    `blocked ${runbook.channel}:${runbook.status} step ${stepIndex + 1}: waiting for channel owner`
  );

  const recoverRunbookCheck = async (runbook: ChannelRunbook, step: string, stepIndex: number) => updateRunbookCheckStatus(
    runbook,
    step,
    stepIndex,
    '/api/ops/channel-runbook-checks/recover',
    `recovered ${runbook.channel}:${runbook.status} step ${stepIndex + 1}`
  );

  const dispatchChannelNotification = async (notification: ChannelNotification, outcome: 'SUCCESS' | 'webhook_timeout') => {
    setError('');
    try {
      await api<ChannelNotification>('/api/ops/channel-notifications/dispatch', {
        method: 'POST',
        body: JSON.stringify({ id: notification.id, outcome })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'dispatch notification failed');
    }
  };

  const updateChannelAlertPolicy = async (policy: ChannelAlertPolicy) => {
    const draft = notificationPolicyDrafts[policy.channel];
    if (!draft) {
      return;
    }
    setError('');
    try {
      await api<ChannelAlertPolicy>('/api/ops/channel-alert-policies/update', {
        method: 'POST',
        body: JSON.stringify({
          channel: policy.channel,
          targetUrl: draft.targetUrl,
          secretRef: draft.secretRef,
          maxAttempts: Number(draft.maxAttempts),
          backoffSeconds: Number(draft.backoffSeconds),
          inboundAuditMinSamples: Number(draft.inboundAuditMinSamples),
          inboundAuditMinAcceptanceRate: Number(draft.inboundAuditMinAcceptanceRate),
          inboundAuditMaxErrorCount: Number(draft.inboundAuditMaxErrorCount),
          actor: 'ops-a',
          note: '控制台更新通知与验收阈值配置'
        })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'update notification policy failed');
    }
  };

  const rollbackChannelAlertPolicy = async (policy: ChannelAlertPolicy) => {
    const confirmation = `ROLLBACK ${policy.channel}`;
    if (!window.confirm(`确认回滚 ${policy.channel} 通知目标？审批人需要提交确认短语：${confirmation}`)) {
      return;
    }
    setError('');
    try {
      await api<ChannelAlertPolicy>('/api/ops/channel-alert-policies/rollback', {
        method: 'POST',
        body: JSON.stringify({ channel: policy.channel, actor: 'ops-lead', note: '通知目标回滚到上一版', confirmation })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'rollback notification policy failed');
    }
  };

  const approveNotificationPolicyChange = async (change: NotificationPolicyChange) => {
    if (!window.confirm(`确认批准 ${change.channel} 通知目标变更？审批人需要提交确认短语：${change.confirmationText}`)) {
      return;
    }
    setError('');
    try {
      await api<ChannelAlertPolicy>('/api/ops/channel-alert-policies/approve-change', {
        method: 'POST',
        body: JSON.stringify({ id: change.id, approver: 'ops-lead', note: '通知目标审批通过', confirmation: change.confirmationText })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'approve notification policy change failed');
    }
  };

  const rejectNotificationPolicyChange = async (change: NotificationPolicyChange) => {
    setError('');
    try {
      await api<NotificationPolicyChange>('/api/ops/channel-alert-policies/reject-change', {
        method: 'POST',
        body: JSON.stringify({ id: change.id, reviewer: 'ops-lead', note: '通知目标未完成备案' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'reject notification policy change failed');
    }
  };

  const cancelNotificationPolicyChange = async (change: NotificationPolicyChange) => {
    setError('');
    try {
      await api<NotificationPolicyChange>('/api/ops/channel-alert-policies/cancel-change', {
        method: 'POST',
        body: JSON.stringify({ id: change.id, actor: 'ops-a', note: '申请人撤销通知目标变更' })
      });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'cancel notification policy change failed');
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
      await refreshInboundAudits();
      await loadMessages(payload.data.conversation.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'channel demo failed');
    } finally {
      setChannelDemoSending('');
    }
  };

  const sendChannelFailureDemo = async (errorExample: ChannelErrorExample) => {
    const example = protocolExamples.find((item) => item.id === errorExample.exampleId);
    if (!example) {
      setError(`missing channel example ${errorExample.exampleId}`);
      return;
    }
    setError('');
    setChannelFailureSending(errorExample.id);
    setChannelFailureResult(null);
    try {
      const headers = errorHeaders(example, errorExample);
      let response: Response | null = null;
      let payload = {} as ApiResponse<unknown>;
      if (errorExample.mutation === 'volume') {
        for (let idx = 0; idx < 70; idx += 1) {
          const volumeRequest = await signedDemoRequest(example);
          response = await fetch(endpointPath(example.endpoint), {
            method: 'POST',
            headers,
            body: JSON.stringify(volumeRequest)
          });
          payload = (await response.json()) as ApiResponse<unknown>;
          if (response.status === errorExample.status) {
            break;
          }
        }
      } else {
        const request = await failedDemoRequest(example, errorExample);
        if (errorExample.mutation === 'duplicate') {
          await fetch(endpointPath(example.endpoint), {
            method: 'POST',
            headers,
            body: JSON.stringify(request)
          });
        }
        response = await fetch(endpointPath(example.endpoint), {
          method: 'POST',
          headers,
          body: JSON.stringify(request)
        });
        payload = (await response.json()) as ApiResponse<unknown>;
      }
      if (!response) {
        throw new Error('channel failure demo did not receive a response');
      }
      const code = payload.error?.code ?? `http_${response.status}`;
      if (response.status !== errorExample.status || code !== errorExample.code) {
        throw new Error(`expected ${errorExample.status} ${errorExample.code}, got ${response.status} ${code}`);
      }
      setChannelFailureResult({
        exampleId: errorExample.id,
        status: response.status,
        code,
        reason: errorExample.reason
      });
      await load();
      await refreshInboundAudits();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'channel failure demo failed');
    } finally {
      setChannelFailureSending('');
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

  const downloadChannelOpsReport = async (format: 'markdown' | 'csv') => {
    const response = await fetch(`/api/ops/channel-ops-report/export?format=${format}`);
    if (!response.ok) {
      throw new Error('download channel ops report failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-ops-${new Date().toISOString().slice(0, 10)}.${format === 'csv' ? 'csv' : 'md'}`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const downloadChannelOpsReportEvents = async () => {
    const response = await fetch(`/api/ops/channel-ops-report-events/export?${channelOpsEventQuery(reportEventStatusFilter, reportEventActorFilter, 50)}`);
    if (!response.ok) {
      throw new Error('download channel ops report events failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-ops-events-${new Date().toISOString().slice(0, 10)}.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const downloadChannelRunbookChecks = async () => {
    const response = await fetch(`/api/ops/channel-runbook-checks/export?${channelRunbookCheckQuery(runbookCheckChannelFilter, runbookCheckStatusFilter, runbookCheckActorFilter, 100)}`);
    if (!response.ok) {
      throw new Error('download channel runbook checks failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-runbook-checks-${new Date().toISOString().slice(0, 10)}.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const downloadChannelInboundAudits = async () => {
    const response = await fetch(`/api/ops/channel-inbound-audits/export?${channelInboundAuditQuery(inboundAuditChannelFilter, inboundAuditStatusFilter, inboundAuditCodeFilter, 100)}`);
    if (!response.ok) {
      throw new Error('download channel inbound audits failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-inbound-audits-${new Date().toISOString().slice(0, 10)}.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const downloadChannelInboundAuditQualityEvents = async () => {
    const response = await fetch(`/api/ops/channel-inbound-audit-quality-events/export?${channelInboundAuditQualityEventQuery(inboundAuditChannelFilter, auditQualityEventStatusFilter, inboundAuditCodeFilter, 100)}`);
    if (!response.ok) {
      throw new Error('download channel inbound audit quality events failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-audit-quality-events-${new Date().toISOString().slice(0, 10)}.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const generateChannelOpsReport = async (format: 'markdown' | 'csv') => {
    setReportGenerating(format);
    setError('');
    try {
      const report = await api<ChannelOpsReport>('/api/ops/channel-ops-reports/generate', {
        method: 'POST',
        body: JSON.stringify({ format })
      });
      setChannelOpsReports((current) => [report, ...current.filter((item) => item.id !== report.id)].slice(0, 6));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'generate channel ops report failed');
    } finally {
      setReportGenerating('');
    }
  };

  const compensateChannelOpsReport = async () => {
    setReportCompensating(true);
    setError('');
    try {
      const result = await api<ChannelOpsReportCompensationResult>('/api/ops/channel-ops-report-scheduler/compensate', {
        method: 'POST',
        body: JSON.stringify({ actor: 'ops-lead', note: '控制台补生成渠道运营日报' })
      });
      setReportScheduler(result.status);
      await refreshReportEvents();
      if (result.report) {
        setChannelOpsReports((current) => [result.report as ChannelOpsReport, ...current.filter((item) => item.id !== result.report?.id)].slice(0, 6));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'compensate channel ops report failed');
    } finally {
      setReportCompensating(false);
    }
  };

  const downloadSavedChannelOpsReport = async (report: ChannelOpsReport) => {
    const response = await fetch(`/api/ops/channel-ops-reports/export?id=${encodeURIComponent(report.id)}`);
    if (!response.ok) {
      throw new Error('download saved channel ops report failed');
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `agent-customer-service-channel-ops-${report.generatedAt.slice(0, 10)}.${report.format === 'csv' ? 'csv' : 'md'}`;
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
          <article className="metric">
            <span>QA queue</span>
            <strong>{openReviewCount}</strong>
            <small>assistant replies awaiting review</small>
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
                    {signatureProfiles.find((profile) => profile.adapterEndpoint === example.endpoint) && (
                      <span>
                        {signatureProfiles.find((profile) => profile.adapterEndpoint === example.endpoint)?.signatureHeader}
                        {' · '}
                        {signatureProfiles.find((profile) => profile.adapterEndpoint === example.endpoint)?.replayHeader}
                      </span>
                    )}
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
                    {signatureProfiles.find((profile) => profile.adapterEndpoint === example.endpoint) && (
                      <span>sig {signatureProfiles.find((profile) => profile.adapterEndpoint === example.endpoint)?.sampleSignature.slice(0, 12)}...</span>
                    )}
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
                    <span>{row.signatureHeader} / {row.timestampHeader}</span>
                  </div>
                  <div>
                    <span>{row.origin}</span>
                    <span>{row.replayKey}</span>
                    <span>{row.rateLimit}</span>
                    <span>{row.retrySemantics}</span>
                  </div>
                  <div>
                    <span>{row.contentField} to content</span>
                    <span>{row.customerField} to customer</span>
                    <span>{signatureProfiles.find((profile) => profile.id === row.signatureProfileId)?.canonicalPayload.join(' + ')}</span>
                    <span>{row.errors.slice(0, 3).join(' / ')}</span>
                  </div>
                </article>
              ))}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">失败演示</p>
                <h2>拒绝边界</h2>
              </div>
              <span className="status warning">{errorExamples.length}</span>
            </div>
            <div className="failureList">
              {errorExamples.map((item) => (
                <article className="failureExample" key={item.id}>
                  <div>
                    <strong>{item.code}</strong>
                    <span>{item.mutation} · {item.reason}</span>
                    {channelFailureResult?.exampleId === item.id && <small>{channelFailureResult.status} · {channelFailureResult.code}</small>}
                  </div>
                  <button
                    className="tinyButton"
                    onClick={() => sendChannelFailureDemo(item)}
                    title="发送失败演示"
                    disabled={channelFailureSending === item.id}
                  >
                    <Send size={14} />
                  </button>
                </article>
              ))}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">验收审计</p>
                <h2>最近入站</h2>
              </div>
              <span className="status">{channelInboundAudits.length}</span>
            </div>
            <div className="eventFilters">
              {auditStatusOptions.map((item) => (
                <button
                  className={inboundAuditStatusFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => {
                    setInboundAuditStatusFilter(item);
                    void refreshInboundAudits(inboundAuditChannelFilter, item, inboundAuditCodeFilter).catch((err) => setError(err instanceof Error ? err.message : 'load channel inbound audits failed'));
                  }}
                >
                  {item}
                </button>
              ))}
              <select
                aria-label="验收审计渠道"
                value={inboundAuditChannelFilter}
                onChange={(event) => {
                  setInboundAuditChannelFilter(event.target.value);
                  void refreshInboundAudits(event.target.value, inboundAuditStatusFilter, inboundAuditCodeFilter).catch((err) => setError(err instanceof Error ? err.message : 'load channel inbound audits failed'));
                  void refreshAuditQualityEvents(event.target.value, auditQualityEventStatusFilter, inboundAuditCodeFilter).catch((err) => setError(err instanceof Error ? err.message : 'load audit quality events failed'));
                }}
              >
                <option value="ALL">ALL</option>
                {auditChannels.map((channel) => (
                  <option value={channel} key={channel}>{channel}</option>
                ))}
              </select>
              <input
                aria-label="验收审计错误码"
                value={inboundAuditCodeFilter}
                onBlur={(event) => {
                  void refreshInboundAudits(inboundAuditChannelFilter, inboundAuditStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load channel inbound audits failed'));
                  void refreshAuditQualityEvents(inboundAuditChannelFilter, auditQualityEventStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load audit quality events failed'));
                }}
                onChange={(event) => setInboundAuditCodeFilter(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    void refreshInboundAudits(inboundAuditChannelFilter, inboundAuditStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load channel inbound audits failed'));
                    void refreshAuditQualityEvents(inboundAuditChannelFilter, auditQualityEventStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load audit quality events failed'));
                  }
                }}
                placeholder="code"
              />
              <button className="tinyButton" onClick={() => downloadChannelInboundAudits().catch((err) => setError(err instanceof Error ? err.message : 'download channel inbound audits failed'))} title="导出验收审计">
                <Download size={14} />
              </button>
            </div>
            {auditQualityRows.length > 0 && (
              <div className="auditQualityGrid">
                {auditQualityRows.map((item) => (
                  <article className="auditQualityRow" key={item.channel}>
                    <div>
                      <strong>{item.channel}</strong>
                      <span>{item.accepted}/{item.total} accepted · {item.acceptanceRate}%</span>
                    </div>
                    <div className="auditQualityBar">
                      <span style={{ width: `${item.acceptanceRate}%` }} />
                      <em style={{ width: `${100 - item.acceptanceRate}%` }} />
                    </div>
                    <b className={item.rejected > 0 ? 'status warning' : 'status'}>{item.topError || 'clean'}</b>
                  </article>
                ))}
              </div>
            )}
            <div className="eventFilters">
              {auditQualityEventStatusOptions.map((item) => (
                <button
                  className={auditQualityEventStatusFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => {
                    setAuditQualityEventStatusFilter(item);
                    void refreshAuditQualityEvents(inboundAuditChannelFilter, item, inboundAuditCodeFilter).catch((err) => setError(err instanceof Error ? err.message : 'load audit quality events failed'));
                  }}
                >
                  {item}
                </button>
              ))}
              <button className="tinyButton" onClick={() => downloadChannelInboundAuditQualityEvents().catch((err) => setError(err instanceof Error ? err.message : 'download audit quality events failed'))} title="导出验收质量事件">
                <Download size={14} />
              </button>
            </div>
            {auditRecoveryRows.length > 0 && (
              <div className="auditRecoveryGrid">
                {auditRecoveryRows.map((row) => (
                  <article className="auditRecoveryRow" key={row.channel}>
                    <div>
                      <strong>{row.channel} · {row.state}</strong>
                      <span>{row.failureCode} · last {row.eventRate}% · current {row.currentRate}%</span>
                    </div>
                    <div className="auditQualityBar">
                      <span style={{ width: `${Math.min(100, row.currentRate)}%` }} />
                      <em style={{ width: `${Math.max(0, 100 - row.currentRate)}%` }} />
                    </div>
                    <small>{row.currentTotal}/{row.minSamples} samples · min {row.minAcceptanceRate}% · same-code {row.maxErrorCount}</small>
                  </article>
                ))}
              </div>
            )}
            {auditQualityEvents.length > 0 && (
              <div className="eventStrip">
                {auditQualityEvents.map((event) => (
                  <span className={event.status === 'ESCALATE' ? 'status danger' : 'status warning'} key={event.id}>
                    {event.channel} · {event.status} · {event.failureCode} · {event.accepted}/{event.total} accepted · {event.acceptanceRate}% · {event.createdAt.slice(5, 16).replace('T', ' ')}
                  </span>
                ))}
              </div>
            )}
            <div className="tableList compactList">
              {channelInboundAudits.slice(0, 6).map((audit) => (
                <article className="tableRow" key={audit.id}>
                  <div>
                    <strong>{audit.channel} · {audit.status}</strong>
                    <span>{audit.code} · {audit.reason || 'channel inbound accepted'}</span>
                    <span>{[audit.origin, audit.externalConversationId, audit.externalMessageId].filter(Boolean).join(' / ')}</span>
                    {audit.signaturePreview && <span>sig {audit.signaturePreview}... · {audit.replayKey?.slice(0, 12)}...</span>}
                  </div>
                  <b className={audit.status === 'REJECTED' ? 'status danger' : 'status'}>{audit.createdAt.slice(11, 19)}</b>
                </article>
              ))}
              {channelInboundAudits.length === 0 && <p className="empty">暂无渠道验收审计</p>}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">失败指标</p>
                <h2>最近 24 小时</h2>
              </div>
              <span className="status warning">{channelAlerts.length}</span>
            </div>
            <div className="alertList">
              {channelAlerts.map((alert) => (
                <article className="alertRow" key={`${alert.channel}-${alert.code}`}>
                  <div>
                    <strong>{alert.channel}</strong>
                    <span>{alert.code} · {alert.lastReason}</span>
                    {alert.lastOrigin && <span>{alert.lastOrigin}</span>}
                  </div>
                  <div>
                    <em>{alert.count}</em>
                    <b className="status warning">{alert.lastSeenAt.slice(11, 19)}</b>
                  </div>
                </article>
              ))}
              {channelAlerts.length === 0 && <p className="empty">暂无渠道失败记录</p>}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">失败趋势</p>
                <h2>小时桶</h2>
              </div>
              <div className="headerActions">
                <span className="status">{channelFailureTrends.length}</span>
                <button className="tinyButton" onClick={() => downloadChannelOpsReport('markdown').catch((err) => setError(err instanceof Error ? err.message : 'download channel ops report failed'))} title="导出 Markdown 日报">
                  <Download size={14} />
                </button>
                <button className="tinyButton" onClick={() => downloadChannelOpsReport('csv').catch((err) => setError(err instanceof Error ? err.message : 'download channel ops report failed'))} title="导出 CSV 日报">
                  <Download size={14} />
                </button>
              </div>
            </div>
            <div className="trendGrid">
              {channelFailureTrends.slice(0, 8).map((item) => (
                <article className="trendCell" key={`${item.channel}-${item.bucketStart}`}>
                  <div>
                    <strong>{item.channel}</strong>
                    <span>{item.bucketStart.slice(11, 16)}</span>
                  </div>
                  <div className="trendBar">
                    <span style={{ width: `${Math.max(8, Math.round((item.count / maxTrendCount) * 100))}%` }} />
                  </div>
                  <b>{item.count}</b>
                </article>
              ))}
              {channelFailureTrends.length === 0 && <p className="empty">暂无失败趋势</p>}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">日报留存</p>
                <h2>渠道运营</h2>
              </div>
              <div className="headerActions">
                <button className="tinyButton" disabled={reportGenerating === 'markdown'} onClick={() => void generateChannelOpsReport('markdown')} title="生成 Markdown 日报">
                  <Save size={14} />
                </button>
                <button className="tinyButton" disabled={reportGenerating === 'csv'} onClick={() => void generateChannelOpsReport('csv')} title="生成 CSV 日报">
                  <Save size={14} />
                </button>
                <button className="tinyButton" disabled={reportCompensating} onClick={() => void compensateChannelOpsReport()} title="补生成日报">
                  <RefreshCcw size={14} />
                </button>
              </div>
            </div>
            <div className="schedulerStrip">
              <span className={reportScheduler?.enabled ? 'status' : 'status muted'}>{reportScheduler?.enabled ? 'Scheduler on' : 'Scheduler off'}</span>
              <span>{reportScheduler?.format ?? 'markdown'} · every {reportScheduler?.intervalMins ?? 1440}m</span>
              <span>retain {reportScheduler?.retain ?? 30}</span>
              <span>{reportScheduler?.lastStatus ?? 'DISABLED'}</span>
              {reportScheduler?.lastRunAt && <span>last {reportScheduler.lastRunAt.slice(5, 16).replace('T', ' ')}</span>}
              {reportScheduler?.nextRunAt && <span>next {reportScheduler.nextRunAt.slice(5, 16).replace('T', ' ')}</span>}
              {reportScheduler?.lastReportId && <span>{reportScheduler.lastReportId}</span>}
              {reportScheduler?.lastError && <span className="dangerText">{reportScheduler.lastError}</span>}
            </div>
            <div className="eventFilters">
              {reportEventStatusOptions.map((item) => (
                <button
                  className={reportEventStatusFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => {
                    setReportEventStatusFilter(item);
                    void refreshReportEvents(item, reportEventActorFilter).catch((err) => setError(err instanceof Error ? err.message : 'load channel ops report events failed'));
                  }}
                >
                  {item}
                </button>
              ))}
              <input
                aria-label="补偿事件操作者"
                value={reportEventActorFilter}
                onBlur={(event) => void refreshReportEvents(reportEventStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load channel ops report events failed'))}
                onChange={(event) => setReportEventActorFilter(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    void refreshReportEvents(reportEventStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load channel ops report events failed'));
                  }
                }}
                placeholder="actor"
              />
              <button className="tinyButton" onClick={() => downloadChannelOpsReportEvents().catch((err) => setError(err instanceof Error ? err.message : 'download channel ops report events failed'))} title="导出补偿事件">
                <Download size={14} />
              </button>
            </div>
            {reportEvents.length > 0 && (
              <div className="eventStrip">
                {reportEvents.map((event) => (
                  <span className={event.status === 'FAILED' ? 'status danger' : 'status'} key={event.id}>
                    {event.action} · {event.actor} · {event.status} · {event.createdAt.slice(5, 16).replace('T', ' ')}
                  </span>
                ))}
              </div>
            )}
            <div className="tableList compactList">
              {channelOpsReports.map((report) => {
                const handoffActions = (report.summary.handoffPriorities ?? [])
                  .filter((item, index, list) => item.notificationId && list.findIndex((next) => next.notificationId === item.notificationId) === index)
                  .slice(0, 3);
                return (
                  <article className="tableRow" key={report.id}>
                    <div>
                      <strong>{report.format.toUpperCase()} · {report.generatedAt.slice(0, 16).replace('T', ' ')}</strong>
                      <span>{report.summary.failureCount} failures · {report.summary.activeRunbooks} runbooks · {report.summary.openNotifications + report.summary.retrying} open notices</span>
                      {report.summary.inboundAudit && (
                        <span>
                          inbound {report.summary.inboundAudit.accepted}/{report.summary.inboundAudit.total} accepted · {report.summary.inboundAudit.acceptanceRate}%
                          {(report.summary.inboundAudit.topErrorCodes ?? []).length > 0 ? ` · ${(report.summary.inboundAudit.topErrorCodes ?? []).map((item) => `${item.code}:${item.count}`).join(' / ')}` : ''}
                        </span>
                      )}
                      {report.summary.inboundAuditQuality && (
                        <span>
                          quality events {report.summary.inboundAuditQuality.eventCount} · active {report.summary.inboundAuditQuality.active} · watch {report.summary.inboundAuditQuality.watch} · recovered {report.summary.inboundAuditQuality.recovered}
                          {(report.summary.inboundAuditQuality.activeChannels ?? []).length > 0 ? ` · ${report.summary.inboundAuditQuality.activeChannels?.join(' / ')}` : ''}
                        </span>
                      )}
                      {(report.summary.handoffPriorities ?? []).length > 0 && (
                        <span>
                          handoff {(report.summary.handoffPriorities ?? []).slice(0, 3).map((item) => `#${item.rank} ${item.channel} ${item.source} · ${item.actionLabel ?? item.actionType ?? item.source}`).join(' / ')}
                        </span>
                      )}
                      <span>{(report.summary.channels ?? []).length > 0 ? (report.summary.channels ?? []).join(' / ') : 'ALL channels'}</span>
                    </div>
                    <div className="gapActions">
                      {handoffActions.map((item) => (
                        <button
                          className="tinyButton"
                          disabled={item.notificationId ? acknowledgedNotificationIds.has(item.notificationId) : true}
                          key={`${report.id}-${item.rank}-${item.notificationId}`}
                          onClick={() => void acknowledgeHandoffPriority(report, item)}
                          title={item.notificationId && acknowledgedNotificationIds.has(item.notificationId) ? '交接通知已确认' : item.actionLabel ?? '确认交接通知'}
                        >
                          <CheckCircle2 size={14} />
                        </button>
                      ))}
                      <button className="tinyButton" onClick={() => downloadSavedChannelOpsReport(report).catch((err) => setError(err instanceof Error ? err.message : 'download saved channel ops report failed'))} title="下载历史日报">
                        <Download size={14} />
                      </button>
                    </div>
                  </article>
                );
              })}
              {channelOpsReports.length === 0 && <p className="empty">暂无历史日报</p>}
            </div>
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">通知策略</p>
                <h2>失败告警</h2>
              </div>
              <span className="status warning">{channelAlertPolicies.filter((item) => item.active).length}</span>
            </div>
            <div className="tableList">
              {channelAlertPolicies.map((policy) => {
                const draft = notificationPolicyDrafts[policy.channel] ?? {
                  targetUrl: policy.targetUrl,
                  secretRef: policy.secretRef,
                  maxAttempts: policy.maxAttempts,
                  backoffSeconds: policy.backoffSeconds,
                  inboundAuditMinSamples: policy.inboundAuditMinSamples,
                  inboundAuditMinAcceptanceRate: policy.inboundAuditMinAcceptanceRate,
                  inboundAuditMaxErrorCount: policy.inboundAuditMaxErrorCount
                };
                return (
                  <article className="tableRow" key={policy.channel}>
                    <div>
                      <strong>{policy.channel} · {policy.notifyTarget}</strong>
                      <span>{policy.currentCount}/{policy.threshold} in {policy.windowMinutes}m</span>
                      <span>{policy.targetUrl} · {policy.secretRef}</span>
                      <span>{policy.maxAttempts} attempts · {policy.backoffSeconds}s base backoff</span>
                      <span>inbound audit: {policy.inboundAuditMinSamples} samples · {policy.inboundAuditMinAcceptanceRate}% min · {policy.inboundAuditMaxErrorCount} same-code</span>
                      {policy.lastTriggeredAt && <span>{policy.lastTriggeredAt.slice(11, 19)}</span>}
                      <div className="policyConfig">
                        <input
                          aria-label={`${policy.channel} target URL`}
                          value={draft.targetUrl}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, targetUrl: event.target.value }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} secret ref`}
                          value={draft.secretRef}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, secretRef: event.target.value }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} max attempts`}
                          min={1}
                          max={10}
                          type="number"
                          value={draft.maxAttempts}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, maxAttempts: Number(event.target.value) }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} backoff seconds`}
                          min={1}
                          max={3600}
                          type="number"
                          value={draft.backoffSeconds}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, backoffSeconds: Number(event.target.value) }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} inbound audit min samples`}
                          min={1}
                          max={100}
                          type="number"
                          value={draft.inboundAuditMinSamples}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, inboundAuditMinSamples: Number(event.target.value) }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} inbound audit min acceptance rate`}
                          min={1}
                          max={100}
                          type="number"
                          value={draft.inboundAuditMinAcceptanceRate}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, inboundAuditMinAcceptanceRate: Number(event.target.value) }
                          }))}
                        />
                        <input
                          aria-label={`${policy.channel} inbound audit max error count`}
                          min={1}
                          max={100}
                          type="number"
                          value={draft.inboundAuditMaxErrorCount}
                          onChange={(event) => setNotificationPolicyDrafts((current) => ({
                            ...current,
                            [policy.channel]: { ...draft, inboundAuditMaxErrorCount: Number(event.target.value) }
                          }))}
                        />
                      </div>
                    </div>
                    <div className="gapActions">
                      <em>{policy.severity}</em>
                      <b className={statusClass(policy.active ? 'HIGH' : 'LOW')}>{policy.active ? 'ACTIVE' : 'READY'}</b>
                      <button className="tinyButton" onClick={() => updateChannelAlertPolicy(policy)} title="保存通知和验收阈值">
                        <Save size={14} />
                      </button>
                      <button className="tinyButton" onClick={() => rollbackChannelAlertPolicy(policy)} title="回滚通知目标">
                        <RefreshCcw size={14} />
                      </button>
                    </div>
                  </article>
                );
              })}
              {channelAlertPolicies.length === 0 && <p className="empty">暂无通知策略</p>}
            </div>
            {notificationPolicyChanges.filter((change) => change.status === 'PENDING').length > 0 && (
              <div className="policyEvents">
                {notificationPolicyChanges.filter((change) => change.status === 'PENDING').slice(0, 4).map((change) => (
                  <article className="policyEvent" key={change.id}>
                    <div>
                      <strong>{change.channel} · {change.status}</strong>
                      <span>{change.targetUrl} · {change.secretRef}</span>
                      <small>{change.maxAttempts} attempts · {change.backoffSeconds}s backoff</small>
                      <small>{change.requestedBy} · expires {change.expiresAt.slice(11, 19)}</small>
                      <small>confirm: {change.confirmationText}</small>
                      <small>{change.note}</small>
                      {change.diff.length > 0 && (
                        <div className="policyDiffs">
                          {change.diff.map((item) => (
                            <small key={`${change.id}-${item.field}`}>{item.field}: {item.before}{' -> '}{item.after}</small>
                          ))}
                        </div>
                      )}
                    </div>
                    <div className="gapActions">
                      <button className="tinyButton" onClick={() => approveNotificationPolicyChange(change)} title="批准通知目标变更">
                        <CheckCircle2 size={14} />
                      </button>
                      <button className="tinyButton" onClick={() => cancelNotificationPolicyChange(change)} title="撤销通知目标变更">
                        <RefreshCcw size={14} />
                      </button>
                      <button className="tinyButton dangerButton" onClick={() => rejectNotificationPolicyChange(change)} title="拒绝通知目标变更">
                        <CircleX size={14} />
                      </button>
                    </div>
                  </article>
                ))}
              </div>
            )}
            {notificationPolicyEvents.length > 0 && (
              <div className="policyEvents">
                {notificationPolicyEvents.slice(0, 4).map((event) => (
                  <article className="policyEvent" key={event.id}>
                    <div>
                      <strong>{event.channel} · {event.action}</strong>
                      <span>{event.after}</span>
                      <small>{event.before}</small>
                      <small>{event.note}</small>
                    </div>
                    <b className="status">{event.actor} · {event.createdAt.slice(11, 19)}</b>
                  </article>
                ))}
              </div>
            )}
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">处置 Runbook</p>
                <h2>失败处理</h2>
              </div>
              <span className="status warning">{channelRunbooks.length}</span>
            </div>
            <div className="runbookList">
              {channelRunbooks.slice(0, 4).map((item) => (
                <article className="runbookItem" key={`${item.channel}-${item.status}-${item.failureCode}`}>
                  <div>
                    <strong>{item.channel} · {item.status}</strong>
                    <span>{item.failureCode} · {item.owner}</span>
                    <small>{item.nextAction}</small>
                    <small>{item.escalation}</small>
                    <div className="runbookSteps">
                      {item.steps.slice(0, 4).map((step, stepIndex) => {
                        const completed = (item.checks ?? []).find((check) => check.stepIndex === stepIndex);
                        return (
                          <small key={`${item.channel}-${item.status}-${stepIndex}`}>
                            {completed ? completed.checkStatus.toLowerCase() : 'todo'} · {step}
                            {completed ? ` · ${completed.assignee || completed.actor}` : ''}
                            {!completed && (
                              <button
                                className="tinyButton inlineIconButton"
                                onClick={() => void completeRunbookCheck(item, step, stepIndex)}
                                title="确认 Runbook 检查项"
                              >
                                <CheckCircle2 size={12} />
                              </button>
                            )}
                            {!completed && (
                              <button
                                className="tinyButton inlineIconButton dangerButton"
                                onClick={() => void blockRunbookCheck(item, step, stepIndex)}
                                title="标记 Runbook 检查阻塞"
                              >
                                <CircleX size={12} />
                              </button>
                            )}
                            {completed?.checkStatus === 'BLOCKED' && (
                              <button
                                className="tinyButton inlineIconButton"
                                onClick={() => void recoverRunbookCheck(item, step, stepIndex)}
                                title="恢复 Runbook 检查项"
                              >
                                <RefreshCcw size={12} />
                              </button>
                            )}
                          </small>
                        );
                      })}
                    </div>
                  </div>
                  <div className="runbookProgress">
                    <b className={statusClass(item.status === 'ESCALATE' ? 'HIGH' : item.status === 'RETRY' ? 'MEDIUM' : 'LOW')}>
                      {item.notificationState || item.severity}
                    </b>
                    <small>{item.checkSummary.done}/{item.checkSummary.total} done</small>
                    {item.checkSummary.blocked > 0 && <small className="dangerText">{item.checkSummary.blocked} blocked</small>}
                  </div>
                </article>
              ))}
              {channelRunbooks.length === 0 && <p className="empty">暂无处置 Runbook</p>}
            </div>
            <div className="filterRow">
              {['ALL', ...runbookCheckChannels].map((item) => (
                <button
                  className={runbookCheckChannelFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => {
                    setRunbookCheckChannelFilter(item);
                    void refreshRunbookChecks(item, runbookCheckStatusFilter, runbookCheckActorFilter).catch((err) => setError(err instanceof Error ? err.message : 'load runbook checks failed'));
                  }}
                >
                  {item}
                </button>
              ))}
            </div>
            <div className="filterRow">
              {['ALL', ...runbookCheckStatuses].map((item) => (
                <button
                  className={runbookCheckStatusFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => {
                    setRunbookCheckStatusFilter(item);
                    void refreshRunbookChecks(runbookCheckChannelFilter, item, runbookCheckActorFilter).catch((err) => setError(err instanceof Error ? err.message : 'load runbook checks failed'));
                  }}
                >
                  {item}
                </button>
              ))}
              <input
                aria-label="Runbook 检查操作者"
                value={runbookCheckActorFilter}
                onBlur={(event) => void refreshRunbookChecks(runbookCheckChannelFilter, runbookCheckStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load runbook checks failed'))}
                onChange={(event) => setRunbookCheckActorFilter(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    void refreshRunbookChecks(runbookCheckChannelFilter, runbookCheckStatusFilter, event.currentTarget.value).catch((err) => setError(err instanceof Error ? err.message : 'load runbook checks failed'));
                  }
                }}
                placeholder="actor"
              />
              <button className="tinyButton" onClick={() => downloadChannelRunbookChecks().catch((err) => setError(err instanceof Error ? err.message : 'download channel runbook checks failed'))} title="导出 Runbook 检查">
                <Download size={14} />
              </button>
            </div>
            {runbookCheckRows.length > 0 && (
              <div className="eventStrip">
                {runbookCheckRows.slice(0, 8).map((check) => (
                  <span className="status" key={check.id}>
                    {check.channel} · {check.runbookStatus} · {check.checkStatus} · {check.assignee || check.actor} · {check.dueAt ? check.dueAt.slice(5, 16).replace('T', ' ') : check.completedAt.slice(5, 16).replace('T', ' ')}
                  </span>
                ))}
              </div>
            )}
            <div className="panelDivider" />
            <div className="panelHeader compactHeader">
              <div>
                <p className="sectionLabel">通知事件</p>
                <h2>出站治理</h2>
              </div>
              <span className="status warning">{visibleNotifications.length}/{channelNotifications.length}</span>
            </div>
            <div className="notificationSummary">
              <span>OPEN {notificationStats.open}</span>
              <span>RETRYING {notificationStats.retrying}</span>
              <span>DEAD {notificationStats.dead}</span>
              <span>AUDITED {notificationStats.audited}</span>
            </div>
            <div className="filterRow">
              {notificationStatusOptions.map((item) => (
                <button
                  className={notificationStatusFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => setNotificationStatusFilter(item)}
                >
                  {item}
                </button>
              ))}
            </div>
            <div className="filterRow">
              {['ALL', ...notificationChannels].map((item) => (
                <button
                  className={notificationChannelFilter === item ? 'filterButton active' : 'filterButton'}
                  key={item}
                  onClick={() => setNotificationChannelFilter(item)}
                >
                  {item}
                </button>
              ))}
            </div>
            <div className="tableList">
              {visibleNotifications.slice(0, 8).map((item) => {
                const latestAudit = item.deliveryAudit?.[item.deliveryAudit.length - 1];
                const expanded = expandedNotificationId === item.id;
                return (
                  <article className="tableRow" key={item.id}>
                    <div>
                      <strong>{item.channel} · {item.target}</strong>
                      <span>{item.reason}</span>
                      <span>{item.targetUrl} · {item.secretRef}</span>
                      <span>{item.attempts}/{item.maxAttempts} attempts{item.lastDispatchAt ? ` · ${item.lastDispatchAt.slice(11, 19)}` : ''}</span>
                      {item.nextRetryAt && <span>retry at {item.nextRetryAt.slice(11, 19)} · {item.backoffSeconds}s backoff</span>}
                      {item.signature && <span>sig {item.signature.slice(0, 12)}...</span>}
                      {item.receiptStatus && <span>{item.receiptStatus} · {item.receiptBody}</span>}
                      {latestAudit && <span>audit {latestAudit.payloadHash.slice(0, 12)} · {latestAudit.signaturePreview}</span>}
                      {latestAudit && <span>{latestAudit.requestSummary}</span>}
                      {latestAudit && <span>{latestAudit.responseSummary}</span>}
                      {item.lastError && <span>{item.lastError}</span>}
                      {item.deadLetterReason && <span>{item.deadLetterReason}</span>}
                      {item.ackedBy && <span>{item.ackedBy} · {item.ackNote}</span>}
                      {expanded && (
                        <div className="auditTrail">
                          {(item.deliveryAudit ?? []).map((audit) => (
                            <small key={`${item.id}-${audit.attempt}-${audit.createdAt}`}>
                              #{audit.attempt} · {audit.createdAt.slice(11, 19)} · {audit.payloadHash.slice(0, 12)} · {audit.signaturePreview}<br />
                              {audit.requestSummary}<br />
                              {audit.responseSummary}
                            </small>
                          ))}
                          {(item.deliveryAudit?.length ?? 0) === 0 && <small>暂无投递审计记录</small>}
                        </div>
                      )}
                    </div>
                    <div className="gapActions">
                      <b className={statusClass(item.status === 'DEAD_LETTER' ? 'HIGH' : item.status === 'RETRYING' ? 'MEDIUM' : 'LOW')}>{item.status}</b>
                      <button
                        className="tinyButton"
                        onClick={() => setExpandedNotificationId(expanded ? '' : item.id)}
                        title="查看投递审计"
                      >
                        <FileSearch size={14} />
                      </button>
                      {(item.status === 'OPEN' || item.status === 'RETRYING') && (
                        <>
                          <button className="tinyButton" onClick={() => dispatchChannelNotification(item, 'SUCCESS')} title="发送通知">
                            <Send size={14} />
                          </button>
                          <button className="tinyButton" onClick={() => dispatchChannelNotification(item, 'webhook_timeout')} title="模拟失败重试">
                            <RefreshCcw size={14} />
                          </button>
                        </>
                      )}
                      {(item.status === 'OPEN' || item.status === 'RETRYING' || item.status === 'DEAD_LETTER' || item.status === 'SENT') && (
                        <button className="tinyButton" onClick={() => acknowledgeChannelNotification(item)} title="确认告警">
                          <CheckCircle2 size={14} />
                        </button>
                      )}
                    </div>
                  </article>
                );
              })}
              {visibleNotifications.length === 0 && <p className="empty">暂无通知事件</p>}
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
                <p className="sectionLabel">质检任务</p>
                <h2>待复核队列</h2>
              </div>
              <span className="status">{openReviewCount}</span>
            </div>
            <div className="tableList">
              {reviewTasks.slice(0, 6).map((task) => (
                <article className="tableRow" key={task.id}>
                  <div>
                    <strong>{task.reason}</strong>
                    <span>{task.channel} · {task.conversationId}</span>
                    <span>{task.messageId}</span>
                    {task.assignee && <span>{task.assignee}{task.completedAt ? ` · ${task.completedAt.slice(11, 19)}` : ''}</span>}
                  </div>
                  <div className="gapActions">
                    <b className={statusClass(task.priority)}>{task.status}</b>
                    {task.status !== 'COMPLETED' && (
                      <span>
                        <button className="tinyButton" onClick={() => assignReviewTask(task)} title="领取任务">
                          <UserRoundCheck size={14} />
                        </button>
                        <button className="tinyButton" onClick={() => completeReviewTask(task)} title="完成任务">
                          <CheckCircle2 size={14} />
                        </button>
                      </span>
                    )}
                  </div>
                </article>
              ))}
              {reviewTasks.length === 0 && <p className="empty">暂无待复核任务</p>}
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
              <button className="tinyButton" onClick={compareRules} title="灰度对比">
                <FileSearch size={14} />
              </button>
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
            {ruleComparison && (
              <div className="ruleCompare">
                <div>
                  <b className={statusClass(ruleComparison.current.riskLevel)}>{ruleComparison.current.ruleCode ?? 'ALLOW'}</b>
                  <span>current · {ruleComparison.current.action}</span>
                </div>
                <div>
                  <b className={statusClass(ruleComparison.canary.riskLevel)}>{ruleComparison.canary.ruleCode ?? 'ALLOW'}</b>
                  <span>canary · {ruleComparison.canary.action}</span>
                </div>
                <strong className={ruleComparison.changed ? 'status warning' : 'status'}>{ruleComparison.changed ? 'CHANGED' : 'SAME'}</strong>
                <p>{ruleComparison.recommendation}</p>
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
                  <span>{item.version} · {item.stage}</span>
                  <span>{item.hitCount} hits{item.lastHitAt ? ` · ${item.lastHitAt.slice(11, 19)}` : ''}</span>
                  <p>{item.trigger}</p>
                  <div className="ruleActions">
                    {item.stage === 'canary' && item.enabled && !hasApprovedRuleGate(item) && (
                      <button className="tinyButton" onClick={() => approveRuleRelease(item)} title="提交发布审批">
                        <ShieldCheck size={14} />
                      </button>
                    )}
                    {item.stage === 'canary' && item.enabled && hasApprovedRuleGate(item) && (
                      <button className="tinyButton" onClick={() => publishCanaryRule(item)} title="发布灰度规则">
                        <CheckCircle2 size={14} />
                      </button>
                    )}
                    {item.stage === 'active' && item.code === 'CANCEL_RISK_TRANSFER' && (
                      <button className="tinyButton" onClick={() => rollbackRule(item)} title="回滚规则">
                        <RefreshCcw size={14} />
                      </button>
                    )}
                  </div>
                </article>
              ))}
            </div>
            {ruleApprovals.length > 0 && (
              <div className="ruleEvents">
                {ruleApprovals.slice(0, 4).map((approval) => (
                  <article className="eventRow" key={approval.id}>
                    <div>
                      <strong>{approval.status} · {approval.ruleCode}</strong>
                      <span>{approval.sampleCount} samples · {approval.riskLevel} · {approval.approver}</span>
                      <span>{approval.sampleIds.slice(0, 3).join(' / ')}</span>
                      <div className="approvalSamples">
                        {approval.sampleIds.slice(0, 3).map((sampleId) => {
                          const sample = trainingSampleById.get(sampleId);
                          return (
                            <div className="approvalSample" key={sampleId}>
                              <b className={statusClass(sample?.verdict ?? 'REVIEW')}>{sample?.verdict ?? 'PENDING'}</b>
                              <span>{sample ? `${sample.score} · ${sample.channel} · ${sample.reviewer}` : `${sampleId} · 样本待生成`}</span>
                              {sample && <small>{sample.prompt}</small>}
                              {sample && <small>{sample.note}</small>}
                            </div>
                          );
                        })}
                      </div>
                      <span>{approval.note}</span>
                    </div>
                    <b className={statusClass(approval.status === 'APPROVED' ? 'LOW' : 'REVIEW')}>{approval.createdAt.slice(11, 19)}</b>
                  </article>
                ))}
              </div>
            )}
            {ruleEvents.length > 0 && (
              <div className="ruleEvents">
                {ruleEvents.slice(0, 4).map((event) => (
                  <article className="eventRow" key={event.id}>
                    <div>
                      <strong>{event.action} · {event.ruleCode}</strong>
                      <span>{event.version} · {event.actor}</span>
                      <span>{event.note}</span>
                    </div>
                    <b className={statusClass(event.action === 'ROLLBACK' ? 'REVIEW' : 'LOW')}>{event.createdAt.slice(11, 19)}</b>
                  </article>
                ))}
              </div>
            )}
            {ruleObservations.length > 0 && (
              <div className="ruleEvents">
                {ruleObservations.slice(0, 3).map((observation) => (
                  <article className="eventRow releaseObservation" key={`${observation.ruleCode}-${observation.version}`}>
                    <div>
                      <strong>{observation.ruleCode} · {observation.version}</strong>
                      <span>{observation.window} window · {observation.ruleHits} hits · {observation.transferTickets} transfers</span>
                      <div className="observationMetrics">
                        <span>{observation.lowScoreSamples} low-score samples</span>
                        <span>{observation.averageReview || 'n/a'} review avg</span>
                        <span>{observation.rollbackRecommended ? 'rollback suggested' : 'watching'}</span>
                      </div>
                      <span>{observation.recommendation}</span>
                    </div>
                    <b className={statusClass(observation.riskLevel)}>{observation.riskLevel}</b>
                  </article>
                ))}
              </div>
            )}
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
