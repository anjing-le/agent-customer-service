#!/usr/bin/env node

const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');
const examplesPath = path.join(root, 'contracts/examples/channel-protocols.json');
const matrixPath = path.join(root, 'contracts/channel-protocol-matrix.json');
const contractPath = path.join(root, 'contracts/api-contract.json');

const examples = JSON.parse(fs.readFileSync(examplesPath, 'utf8'));
const matrix = JSON.parse(fs.readFileSync(matrixPath, 'utf8'));
const contract = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
const endpointKeys = new Set((contract.endpoints || []).map((endpoint) => `${endpoint.method} ${endpoint.path}`));
const exampleIDs = new Set((examples.examples || []).map((example) => example.id));
const profileIDs = new Set((examples.platformSignatureProfiles || []).map((profile) => profile.id));
const knownErrorCodes = new Set();

for (const endpoint of contract.endpoints || []) {
  for (const error of endpoint.errorResponses || []) {
    for (const code of String(error.code || '').split('|')) {
      knownErrorCodes.add(code);
    }
  }
}

function fail(message) {
  console.error(`check-channel-examples: ${message}`);
  process.exitCode = 1;
}

if (examples.signatureAlgorithm !== 'HMAC-SHA256') {
  fail('signatureAlgorithm must be HMAC-SHA256');
}

const expectedPayload = ['channel', 'externalConversationId', 'timestamp', 'content'];
if (JSON.stringify(examples.canonicalPayload) !== JSON.stringify(expectedPayload)) {
  fail(`canonicalPayload must be ${expectedPayload.join(', ')}`);
}

for (const item of examples.examples || []) {
  if (!endpointKeys.has(item.endpoint)) {
    fail(`${item.id} references unknown endpoint ${item.endpoint}`);
  }
  if (!item.headers || item.headers['Content-Type'] !== 'application/json') {
    fail(`${item.id} must declare application/json content type`);
  }
  if (!item.headers['X-Channel-Origin']) {
    fail(`${item.id} must declare X-Channel-Origin`);
  }
  const input = item.signatureInput || {};
  const payload = [input.channel, input.externalConversationId, input.timestamp, input.content]
    .map((value) => String(value || '').trim())
    .join('\n');
  const expected = crypto.createHmac('sha256', item.demoSecret || '').update(payload).digest('hex');
  const actual = item.request?.signature;
  if (actual !== expected) {
    fail(`${item.id} signature mismatch: expected ${expected}, got ${actual}`);
  }
}

for (const profile of examples.platformSignatureProfiles || []) {
  if (!endpointKeys.has(profile.adapterEndpoint)) {
    fail(`${profile.id} references unknown endpoint ${profile.adapterEndpoint}`);
  }
  for (const field of ['signatureHeader', 'timestampHeader', 'replayHeader']) {
    if (!String(profile[field] || '').startsWith('X-')) {
      fail(`${profile.id} must declare ${field} as an X-* header`);
    }
  }
  if (!Array.isArray(profile.canonicalPayload) || profile.canonicalPayload.length < 4) {
    fail(`${profile.id} must declare a canonicalPayload with at least four fields`);
  }
  if (!Array.isArray(profile.sampleCanonicalPayload) || profile.sampleCanonicalPayload.length !== profile.canonicalPayload.length) {
    fail(`${profile.id} sampleCanonicalPayload must match canonicalPayload length`);
  }
  const payload = profile.sampleCanonicalPayload.map((value) => String(value || '').trim()).join('\n');
  const expected = crypto.createHmac('sha256', profile.demoSecret || '').update(payload).digest('hex');
  if (profile.sampleSignature !== expected) {
    fail(`${profile.id} sampleSignature mismatch: expected ${expected}, got ${profile.sampleSignature}`);
  }
  for (const code of profile.failureCodes || []) {
    if (!knownErrorCodes.has(code)) {
      fail(`${profile.id} references unknown error code ${code}`);
    }
  }
}

for (const error of examples.errorExamples || []) {
  if (!knownErrorCodes.has(error.code)) {
    fail(`${error.id} references unknown error code ${error.code}`);
  }
  if (error.exampleId && !exampleIDs.has(error.exampleId)) {
    fail(`${error.id} references unknown example ${error.exampleId}`);
  }
  if (!['origin', 'signature', 'timestamp', 'volume', 'duplicate'].includes(error.mutation)) {
    fail(`${error.id} uses unsupported mutation ${error.mutation}`);
  }
}

if (JSON.stringify(matrix.signaturePayload) !== JSON.stringify(expectedPayload)) {
  fail(`matrix signaturePayload must be ${expectedPayload.join(', ')}`);
}

for (const row of matrix.rows || []) {
  if (!endpointKeys.has(row.adapterEndpoint)) {
    fail(`${row.channel} matrix references unknown endpoint ${row.adapterEndpoint}`);
  }
  if (!exampleIDs.has(row.successExampleId)) {
    fail(`${row.channel} matrix references unknown example ${row.successExampleId}`);
  }
  if (!profileIDs.has(row.signatureProfileId)) {
    fail(`${row.channel} matrix references unknown signature profile ${row.signatureProfileId}`);
  }
  if (!row.signatureHeader || !row.timestampHeader) {
    fail(`${row.channel} matrix must declare signature and timestamp headers`);
  }
  for (const code of row.errors || []) {
    if (!knownErrorCodes.has(code)) {
      fail(`${row.channel} matrix references unknown error code ${code}`);
    }
  }
}

if (!process.exitCode) {
  console.log('check-channel-examples: ok');
}
