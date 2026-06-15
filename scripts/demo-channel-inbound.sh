#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BASE_URL="${1:-http://localhost:10002}"
EXAMPLE_FILTER="${2:-all}"

BASE_URL="$BASE_URL" EXAMPLE_FILTER="$EXAMPLE_FILTER" node <<'NODE'
const crypto = require('crypto');
const fs = require('fs');
const http = require('http');
const https = require('https');
const path = require('path');

const root = process.cwd();
const baseUrl = process.env.BASE_URL || 'http://localhost:10002';
const exampleFilter = process.env.EXAMPLE_FILTER || 'all';
const examples = JSON.parse(fs.readFileSync(path.join(root, 'contracts/examples/channel-protocols.json'), 'utf8'));

function canonicalSignature(secret, input) {
  const payload = [input.channel, input.externalConversationId, input.timestamp, input.content]
    .map((value) => String(value || '').trim())
    .join('\n');
  return crypto.createHmac('sha256', secret).update(payload).digest('hex');
}

function endpointPath(endpoint) {
  const [, routePath] = endpoint.split(/\s+/);
  return routePath;
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function requestFor(example, index) {
  const timestamp = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
  const unique = `${Date.now()}-${index}`;
  const request = clone(example.request);
  const signatureInput = clone(example.signatureInput);
  signatureInput.timestamp = timestamp;

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

  const secret = process.env[example.secretRef] || example.demoSecret;
  request.signature = canonicalSignature(secret, signatureInput);
  return request;
}

function postJson(url, headers, body) {
  return new Promise((resolve, reject) => {
    const target = new URL(url);
    const client = target.protocol === 'https:' ? https : http;
    const payload = JSON.stringify(body);
    const req = client.request({
      method: 'POST',
      hostname: target.hostname,
      port: target.port,
      path: `${target.pathname}${target.search}`,
      headers: {
        ...headers,
        'Content-Length': Buffer.byteLength(payload)
      }
    }, (res) => {
      let data = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => {
        resolve({ statusCode: res.statusCode, body: data });
      });
    });
    req.on('error', reject);
    req.write(payload);
    req.end();
  });
}

async function main() {
  const selected = examples.examples.filter((example) => exampleFilter === 'all' || example.id === exampleFilter);
  if (selected.length === 0) {
    throw new Error(`No channel example matched "${exampleFilter}"`);
  }

  for (const [index, example] of selected.entries()) {
    const routePath = endpointPath(example.endpoint);
    const url = new URL(routePath, baseUrl).toString();
    const request = requestFor(example, index + 1);
    const result = await postJson(url, example.headers, request);
    const ok = result.statusCode === example.expectedSuccess.status;
    const preview = result.body.replace(/\s+/g, ' ').slice(0, 180);
    console.log(`${ok ? 'ok' : 'fail'} ${example.id} ${result.statusCode} ${url}`);
    console.log(preview);
    if (!ok) {
      process.exitCode = 1;
      return;
    }
  }
}

main().catch((err) => {
  console.error(`demo-channel-inbound: ${err.message}`);
  process.exitCode = 1;
});
NODE
