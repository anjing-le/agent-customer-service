# Agent Customer Service - 智能电商客服系统

## 核心设计

```
用户消息
  → 保存消息(DB)
    → 智能分析(LLM/关键词兜底) → 场景 + 意图 + 情绪
      → 知识检索(商品/活动/FAQ，支持手动选择)
        → 多轮上下文(最近10条历史)
          → 提示词组装 + LLM生成(失败则规则引擎兜底)
            → 工具选择 + 推理过程构建
              → 返回前端(回复 + 卡片 + 推理链)
```

**双引擎降级**：LLM 不可用时自动切换关键词规则引擎，系统永远可用。

## 功能模块

| 模块 | 功能 | 状态 |
|------|------|:----:|
| **对话中心** | 多会话管理、AI 智能回复、推理过程可视化、多轮上下文 | ✅ 真实 |
| **知识中心** | 商品/活动/FAQ 知识管理与检索 | ✅ 真实 |
| **知识中心** | 行业知识/解决方案管理 | 📋 模拟 |
| **场景配置** | 意图/提示词/规则管理 | 📋 模拟 |

> ✅ 真实 = 接入对话链路，📋 模拟 = CRUD 可用但未接入对话链路，预留扩展

## 技术栈

**后端**：Spring Boot 3.4.5 + JPA + MySQL 8 + Redis + Druid | **前端**：Vue 3.5 + Element Plus + Vite + Pinia + TypeScript | **LLM**：OneRouter (OpenAI 兼容)

## 快速开始

### 1. 数据库

```sql
CREATE DATABASE agent_customer_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

> JPA `ddl-auto: update`，启动自动建表，无需手动。

### 2. 后端

```bash
cd backend
cp src/main/resources/application-local.yml.example src/main/resources/application-local.yml
# 编辑 application-local.yml，填入 MySQL 密码和 LLM API Key
mvn spring-boot:run
```

验证：访问 `http://localhost:10002/api/knowledge/overview` 返回 JSON 即成功。

### 3. 前端

```bash
cd frontend
pnpm install
pnpm dev
```

验证：浏览器访问后，左侧菜单「智能客服」→ 对话中心 / 知识中心 / 场景配置。

## 项目结构

```
backend/src/main/java/com/anjing/
├── module/chat/          # 对话中心（ChatService 8步链路 + LlmService）
├── module/knowledge/     # 知识中心（5类知识 CRUD）
├── module/scene/         # 场景配置（意图/提示词/规则）
├── config/               # CORS、Redis、JPA、Druid 等配置
└── model/                # 统一响应、异常、常量

frontend/src/views/customer-service/
├── chat/                 # 对话中心（四栏布局）
├── knowledge/            # 知识中心（5个Tab）
└── scene/                # 场景配置（3个Tab）
```

## API 接口

| 模块 | 接口 | 说明 |
|------|------|------|
| 对话 | `POST /api/chat/session/create` | 创建会话 |
| 对话 | `POST /api/chat/message/send` | 发送消息（核心链路） |
| 知识 | `POST /api/knowledge/{type}/list` | 查询知识列表 |
| 知识 | `POST /api/knowledge/{type}/save` | 新增/编辑知识 |
| 场景 | `POST /api/scene/{type}/list` | 查询配置列表 |
| 场景 | `POST /api/scene/prompt/test` | 测试提示词（真实调 LLM） |

## 环境变量

| 变量 | 说明 |
|------|------|
| `DB_PASSWORD` | MySQL 密码 |
| `LLM_ENABLED` | 是否启用 LLM（默认 false） |
| `LLM_API_KEY` | LLM API Key |
| `LLM_MODEL` | 模型名（默认 gpt-4o-mini） |

> 完整配置见 `application-local.yml.example`

## License

MIT
