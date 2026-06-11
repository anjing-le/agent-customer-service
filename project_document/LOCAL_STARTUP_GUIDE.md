# Local Startup Guide

## Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

开发端口来自 `frontend/.env` 的 `VITE_PORT`，当前为 `20002`。开发环境通过 Vite proxy 转发 `/api` 到 `http://localhost:10002`。

## Backend

```bash
cd backend
mvn spring-boot:run
```

后端当前端口为 `10002`。

本地敏感配置放在：

```text
backend/src/main/resources/application-local.yml
```

该文件被 ignore，不应提交。

## Smoke Checks

```bash
curl http://localhost:10002/api/knowledge/overview
```

```bash
curl -X POST http://localhost:10002/api/chat/session/create \
  -H "Content-Type: application/json" \
  -d '{"userId":"demo","userName":"演示用户","channel":"web"}'
```

## Quality Checks

```bash
./scripts/check-template.sh
./scripts/check-contracts.sh
```
