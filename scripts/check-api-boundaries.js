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

if (platform.responseEnvelope.successCode !== '0') {
  fail('success code must be string "0"');
}

const apiConstants = readText('backend/src/main/java/com/anjing/model/constants/ApiConstants.java');
const frontendApiFiles = [
  'frontend/src/api/apiPaths.ts',
  'frontend/src/api/customer-service/chat.ts',
  'frontend/src/api/customer-service/knowledge.ts',
  'frontend/src/api/customer-service/scene.ts'
].map(readText).join('\n');

for (const boundary of boundaries.boundaries) {
  if (!boundary.basePath.startsWith(boundaries.apiPrefix)) {
    fail(`${boundary.id} basePath must start with ${boundaries.apiPrefix}`);
  }
  if (!apiConstants.includes(boundary.basePath)) {
    fail(`${boundary.id} basePath ${boundary.basePath} missing in ApiConstants`);
  }
  for (const route of boundary.routes || []) {
    if (!apiConstants.includes(route.path)) {
      fail(`${route.path} missing in backend ApiConstants`);
    }
    if (!frontendApiFiles.includes(route.path)) {
      fail(`${route.path} missing in frontend customer-service API files`);
    }
  }
}

if (frontendApiFiles.includes('m1.apifoxmock.com')) {
  fail('runtime API files must not point to Apifox mock');
}

if (!process.exitCode) {
  console.log('check-api-boundaries: ok');
}
