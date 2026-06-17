# Project Document

本目录存放 `agent-customer-service` 的定位、边界、规划和迁移状态。

## 文档

- [PROJECT_CONSTRAINTS.md](./PROJECT_CONSTRAINTS.md)：长期工程约束和防破窗规则。
- [API_CONTRACT_GUIDE.md](./API_CONTRACT_GUIDE.md)：API 响应、分页和迁移规则。
- [SERVICE_BOUNDARY_GUIDE.md](./SERVICE_BOUNDARY_GUIDE.md)：服务/模块边界和未来拆分方向。
- [SCAFFOLD_INHERITANCE.md](./SCAFFOLD_INHERITANCE.md)：本项目如何从 DVSkyFolding 技术基线生长出来。
- [DEMO_FLOW.md](./DEMO_FLOW.md)：课堂演示路径，从客服主链路到 Runbook 和日报交接。
- [DOMAIN_MODEL.md](./DOMAIN_MODEL.md)：可靠客服 Agent 的领域模型和应用端口。
- [ROADMAP.md](./ROADMAP.md)：V1/V2/V3 路线图。
- [STATUS.md](./STATUS.md)：当前阶段状态和验证入口。
- [LOCAL_STARTUP_GUIDE.md](./LOCAL_STARTUP_GUIDE.md)：本地启动和 smoke check。

## 维护原则

- 文档必须区分真实运行链路和预留/模拟能力。
- 教学表达必须先说明脚手架继承关系，再说明客服 Agent 的业务增量。
- 新接口先更新契约，再实现后端和前端。
- 每批迁移保持小步可验证，不为了套模板破坏现有业务演示。
