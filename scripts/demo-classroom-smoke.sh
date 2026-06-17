#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BASE_URL="${1:-http://localhost:10002}"

echo "demo-classroom-smoke: checking ${BASE_URL}"

BASE_URL="$BASE_URL" node <<'NODE'
const http = require('http');
const https = require('https');

const baseUrl = process.env.BASE_URL || 'http://localhost:10002';

function request(method, path, body) {
  return new Promise((resolve, reject) => {
    const target = new URL(path, baseUrl);
    const payload = body ? JSON.stringify(body) : '';
    const client = target.protocol === 'https:' ? https : http;
    const req = client.request({
      method,
      hostname: target.hostname,
      port: target.port,
      path: `${target.pathname}${target.search}`,
      headers: {
        Accept: 'application/json, text/html, text/markdown, text/csv',
        ...(body ? { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(payload) } : {})
      }
    }, (res) => {
      let data = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => {
        resolve({ statusCode: res.statusCode || 0, headers: res.headers, body: data });
      });
    });
    req.on('error', reject);
    if (payload) {
      req.write(payload);
    }
    req.end();
  });
}

function expect(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function parseEnvelope(name, result) {
  expect(result.statusCode >= 200 && result.statusCode < 300, `${name} returned HTTP ${result.statusCode}`);
  let payload;
  try {
    payload = JSON.parse(result.body);
  } catch (err) {
    throw new Error(`${name} did not return JSON`);
  }
  expect(payload.success === true, `${name} envelope was not success`);
  return payload.data;
}

function assetPathsFromHtml(html) {
  const assets = [];
  for (const match of html.matchAll(/(?:src|href)="([^"]+)"/g)) {
    if (match[1].startsWith('/assets/')) {
      assets.push(match[1]);
    }
  }
  return assets;
}

async function main() {
  const health = parseEnvelope('healthz', await request('GET', '/healthz'));
  expect(health.status === 'ok', 'healthz status is not ok');
  console.log(`ok healthz ${health.service}`);

  const html = await request('GET', '/');
  expect(html.statusCode === 200, `console returned HTTP ${html.statusCode}`);
  const assetBodies = [];
  for (const assetPath of assetPathsFromHtml(html.body)) {
    const asset = await request('GET', assetPath);
    expect(asset.statusCode === 200, `console asset ${assetPath} returned HTTP ${asset.statusCode}`);
    assetBodies.push(asset.body);
  }
  const consoleText = [html.body, ...assetBodies].join('\n');
  for (const expected of ['课堂主线', '脚手架基线', '客服主链路', 'RAG 与规则', '渠道验收', 'Runbook', '日报交接']) {
    expect(consoleText.includes(expected), `console bundle missing "${expected}"`);
  }
  console.log('ok console classroom journey');

  const dashboard = parseEnvelope('dashboard', await request('GET', '/api/ops/dashboard'));
  expect(Array.isArray(dashboard.metrics) && dashboard.metrics.length >= 4, 'dashboard metrics are incomplete');
  expect(Array.isArray(dashboard.conversations) && dashboard.conversations.length > 0, 'dashboard conversations are missing');
  expect(Array.isArray(dashboard.channelPolicies) && dashboard.channelPolicies.length >= 4, 'dashboard channel policies are missing');
  expect(Array.isArray(dashboard.integrations) && dashboard.integrations.length >= 4, 'dashboard channel integrations are missing');
  expect(dashboard.quality && typeof dashboard.quality.score === 'number', 'dashboard quality summary is missing');
  console.log(`ok dashboard metrics=${dashboard.metrics.length} conversations=${dashboard.conversations.length} channels=${dashboard.channelPolicies.length}`);

  const articles = parseEnvelope('knowledge articles', await request('GET', '/api/knowledge/articles'));
  expect(Array.isArray(articles) && articles.length >= 3, 'knowledge articles are incomplete');
  expect(articles.some((item) => ['HIGH', 'TRUSTED'].includes(item.trustLevel)), 'trusted knowledge article is missing');
  console.log(`ok knowledge articles=${articles.length}`);

  const messageResult = parseEnvelope('send message', await request('POST', '/api/customer-service/messages', {
    conversationId: 'conv_demo_refund',
    content: '7 天无理由退货的运费怎么计算？'
  }));
  expect(messageResult.agentMessage && messageResult.agentMessage.safe === true, 'agent reply is not marked safe');
  expect(Array.isArray(messageResult.evidence) && messageResult.evidence.length > 0, 'agent reply has no RAG evidence');
  expect(messageResult.evidence[0].retrievalScore > 0, 'agent evidence has no retrieval score');
  expect(typeof messageResult.evidence[0].retrievalReason === 'string' && messageResult.evidence[0].retrievalReason.length > 0, 'agent evidence has no retrieval reason');
  expect(messageResult.agentMessage.trace && messageResult.agentMessage.trace.evidenceCount > 0, 'agent trace did not record evidence');
  console.log(`ok agent reply evidence=${messageResult.evidence.length} score=${messageResult.evidence[0].retrievalScore} engine=${messageResult.agentMessage.engine}`);

  const ruleResult = parseEnvelope('rule test', await request('POST', '/api/ops/rules/test', {
    content: '我已经投诉很多次了，现在必须转人工'
  }));
  expect(ruleResult.matched === true, 'rule test did not match transfer rule');
  expect(String(ruleResult.action).toUpperCase().includes('TRANSFER'), `rule action was ${ruleResult.action}`);
  console.log(`ok rule fallback ${ruleResult.ruleCode} -> ${ruleResult.action}`);

  const report = await request('GET', '/api/ops/channel-ops-report/export?format=markdown');
  expect(report.statusCode === 200, `channel ops report returned HTTP ${report.statusCode}`);
  expect(String(report.headers['content-type'] || '').includes('text/markdown'), 'channel ops report is not markdown');
  expect(report.body.includes('# Agent Customer Service Channel Ops Report'), 'channel ops report title is missing');
  console.log('ok channel ops markdown export');
}

main().catch((err) => {
  console.error(`demo-classroom-smoke: ${err.message}`);
  if (err && (err.code === 'ECONNREFUSED' || err.code === 'ENOTFOUND')) {
    console.error(`demo-classroom-smoke: start the app first, for example: pnpm build:console && go run ./cmd/platform-all`);
  }
  process.exitCode = 1;
});
NODE

./scripts/demo-channel-inbound.sh "$BASE_URL" wechat-adapter-inbound

echo "demo-classroom-smoke: ok"
