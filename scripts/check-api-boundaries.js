#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');

function readJson(relativePath) {
  return JSON.parse(fs.readFileSync(path.join(root, relativePath), 'utf8'));
}

function readText(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8');
}

function fail(message) {
  console.error(`check-api-boundaries: ${message}`);
  process.exitCode = 1;
}

const platform = readJson('contracts/platform-contract.json');
const boundaries = readJson('contracts/service-boundaries.json');
const api = readJson('contracts/api-contract.json');

if (platform.apiPrefix !== '/api') {
  fail('platform apiPrefix must be /api');
}

if (platform.responseEnvelope.successField !== 'success') {
  fail('response envelope must use success boolean');
}

const goFiles = [
  'internal/channels/routes.go',
  'internal/customer/routes.go',
  'internal/knowledge/routes.go',
  'internal/ops/routes.go'
].map(readText).join('\n');
const consoleApp = readText('apps/console/src/main.tsx');
const boundaryRoutes = new Map();

for (const boundary of boundaries.boundaries) {
  if (!boundary.basePath.startsWith(boundaries.apiPrefix)) {
    fail(`${boundary.id} basePath must start with ${boundaries.apiPrefix}`);
  }
  for (const route of boundary.routes || []) {
    for (const method of route.methods || []) {
      boundaryRoutes.set(`${method} ${route.path}`, boundary.id);
    }
    if (!goFiles.includes(route.path)) {
      fail(`${route.path} missing in Go route handlers`);
    }
    if (route.console !== false && !consoleApp.includes(route.path)) {
      fail(`${route.path} missing in React console`);
    }
  }
}

if (api.apiPrefix !== boundaries.apiPrefix) {
  fail('api contract prefix must match service boundaries prefix');
}

const contractRoutes = new Set();
for (const endpoint of api.endpoints || []) {
  const key = `${endpoint.method} ${endpoint.path}`;
  contractRoutes.add(key);
  const boundaryID = boundaryRoutes.get(key);
  if (!boundaryID) {
    fail(`${key} missing in service boundaries`);
    continue;
  }
  if (boundaryID !== endpoint.boundary) {
    fail(`${key} boundary mismatch: api=${endpoint.boundary} boundary=${boundaryID}`);
  }
  if (!goFiles.includes(endpoint.path)) {
    fail(`${key} missing in Go route handlers`);
  }
  const requestSchema = endpoint.request?.body || endpoint.request?.query;
  if (requestSchema && !api.schemas?.[requestSchema]) {
    fail(`${key} references missing request schema ${requestSchema}`);
  }
  const responseSchema = normalizeSchemaName(endpoint.response?.schema);
  if (!responseSchema || !api.schemas?.[responseSchema]) {
    fail(`${key} references missing response schema ${endpoint.response?.schema}`);
  }
}

for (const key of boundaryRoutes.keys()) {
  if (!contractRoutes.has(key)) {
    fail(`${key} missing in api contract`);
  }
}

if (!process.exitCode) {
  console.log('check-api-boundaries: ok');
}

function normalizeSchemaName(schema) {
  if (!schema) {
    return '';
  }
  return schema.endsWith('[]') ? schema.slice(0, -2) : schema;
}
