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

if (platform.apiPrefix !== '/api') {
  fail('platform apiPrefix must be /api');
}

if (platform.responseEnvelope.successField !== 'success') {
  fail('response envelope must use success boolean');
}

const goFiles = [
  'internal/customer/routes.go',
  'internal/knowledge/routes.go',
  'internal/ops/routes.go'
].map(readText).join('\n');
const consoleApp = readText('apps/console/src/main.tsx');

for (const boundary of boundaries.boundaries) {
  if (!boundary.basePath.startsWith(boundaries.apiPrefix)) {
    fail(`${boundary.id} basePath must start with ${boundaries.apiPrefix}`);
  }
  for (const route of boundary.routes || []) {
    if (!goFiles.includes(route.path)) {
      fail(`${route.path} missing in Go route handlers`);
    }
    if (route.console !== false && !consoleApp.includes(route.path)) {
      fail(`${route.path} missing in React console`);
    }
  }
}

if (!process.exitCode) {
  console.log('check-api-boundaries: ok');
}
