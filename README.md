# Agent Customer Service - 智能电商客服系统

## 环境要求

| 工具 | 版本 | 必须 |
|------|------|:----:|
| Java | JDK 17+ | ✅ |
| Maven | 3.8+ | ✅ |
| MySQL | 8.0+ | ✅ |
| Node.js | 18+ | ✅ |
| pnpm | 8+ | ✅ |
| Redis | 6.0+ | 可选（项目内置降级） |

## 第一步：创建数据库

```sql
CREATE DATABASE agent_customer_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

> 无需手动建表，JPA 配置了 `ddl-auto: update`，后端启动时自动创建全部 10 张业务表。

## 第二步：配置后端

```bash
cd backend

# 复制示例配置文件
cp src/main/resources/application-local.yml.example src/main/resources/application-local.yml
```

编辑 `application-local.yml`，填入你的配置：

```yaml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/agent_customer_service?useUnicode=true&characterEncoding=utf8&serverTimezone=GMT%2B8&useSSL=false&allowPublicKeyRetrieval=true
    username: root
    password: 你的MySQL密码               # ← 必填

  data:
    redis:
      host: localhost
      port: 6379
      password:                            # ← 如有密码则填入，无密码留空即可

# LLM 大模型配置（可选，不配也能正常对话，走规则引擎兜底）
llm:
  enabled: true                            # ← 改为 true 启用大模型
  api-url: https://api.onerouter.top/v1/chat/completions
  api-key: 你的API_Key                     # ← 从 https://onerouter.top 注册获取
  model: gpt-4o-mini
```

> `application-local.yml` 已被 `.gitignore` 排除，不会上传到 GitHub。

## 第三步：启动后端

```bash
cd backend
mvn clean compile -DskipTests
mvn spring-boot:run
```

或在 IDE 中直接运行 `Application.java` 主类。

**验证后端**：

```bash
curl http://localhost:10002/api/knowledge/overview
```

返回 JSON 数据即表示后端启动成功：

```json
{"code":200,"data":{"productCount":5,"activityCount":3,"faqCount":5,"industryCount":3,"solutionCount":3}}
```

## 第四步：启动前端

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

开发环境已配置代理，`/api` 请求自动转发到 `http://localhost:10002`。

**验证前端**：

浏览器访问后登录系统，左侧菜单「智能客服」下有三个子菜单：

- **对话中心** → 发送消息测试 AI 回复
- **知识中心** → 查看/新增知识数据
- **场景配置** → 查看意图/提示词/规则配置

## 常见问题

### Q：后端启动报 `Communications link failure`

MySQL 连接失败。检查：MySQL 是否启动、端口是否 3306、密码是否正确。

### Q：对话只返回固定文案，LLM 没生效

`llm.enabled` 为 false 或 API Key 无效。在 `application-local.yml` 中设置 `llm.enabled: true` 并填入有效 Key。

> 即使 LLM 不可用，系统也能正常对话——会自动降级为关键词规则引擎。

### Q：`pnpm: command not found`

```bash
npm install -g pnpm
```

### Q：Redis 没装，影响启动吗？

不影响。项目内置了 Redis 不可用时的本地锁降级，核心功能正常运行。

## License

MIT
