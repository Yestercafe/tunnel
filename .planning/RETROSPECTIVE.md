# Project Retrospective

*在里程碑结束时追加；经验用于下一轮规划。*

## Milestone: v1.0 — v1 协议规范与一致性测试

**Shipped:** 2026-04-04  
**Phases:** 6 | **Plans:** 16 | **Tasks:** 17（自小结统计）

### What Was Built

- v1 规范树（`docs/spec/v1/`）：帧、TLS 成帧、会话、路由、流、应用信封、状态机、错误码、安全假设  
- 参考实现包：`pkg/framing`、`pkg/appenvelope`  
- `testdata/` golden/负例与 GitHub Actions 上 `go test ./... -count=1`

### What Worked

- **协议先行 + 每阶段 PLAN/SUMMARY** 使范围可核对，需求与可追溯表可对齐  
- **Go 测试 + hex 向量** 比纯文字更易锁定边界条件

### What Was Inefficient

- **REQUIREMENTS.md 复选框** 曾落后于可追溯表，收尾时需一次性 `requirements mark-complete` 同步  
- 单日主提交集中，里程碑时间线对「逐日速率」参考有限

### Patterns Established

- 帧头 **version / capability**、**`HAS_APP_ENVELOPE`** 与 **`application_data`** 边界的写法与测试约定应延续到实现里程碑

### Key Lessons

1. 里程碑收尾前应对照 **Traceability** 与 **checkbox**，避免「表已 Complete、列表仍空」  
2. 归档后 **ROADMAP / REQUIREMENTS** 应变薄；历史进 `milestones/` 保持主文件可读

### Cost Observations

- 模型与会话占比未计量；后续若启用多模型可补记

---

## Milestone: v1.1 — 最小 Relay 与 Client

**Shipped:** 2026-04-14  
**Phases:** 5（7–11） | **Plans:** 11 | **Tasks:** 24（CLI 统计）

### What Was Built

- **`pkg/protocol`**：PROT-01/02 与 `join_gate` 一致语义  
- **`internal/fakepeer` + `pkg/client`**：TLS 上 SESSION_CREATE/JOIN、JOIN 后 `STREAM_DATA`；`cmd/tunnel client` 冒烟  
- **`pkg/relay` + `cmd/tunnel relay`**：监听、Registry、数据面广播/单播、非法路由与 JOIN 前数据面 `PROTOCOL_ERROR`  
- **`pkg/relay/relay_test.go`**：E2E-01/02 与追溯文档对齐  

### What Worked

- **roadmap analyze**（`disk_status` / 计划与 SUMMARY 计数）便于里程碑收尾前自检  
- 真实 **`relay.Server`** 上的集成测试比仅 fake peer 更能锁住 Relay 侧门禁  

### What Was Inefficient

- 路线图正文与 **Phase 7** 勾选曾滞后（`roadmap_complete: false`），收尾依赖人工对齐归档  
- **Traceability** 曾短期与 **PROT-01/02** 状态不一致，需在归档前校正  

### Patterns Established

- 测试专用 **`client.UnderlyingTLSConn()`** 发送 API 未暴露的帧，用于负例而不放宽生产路径  

### Key Lessons

1. 里程碑完成前跑一次 **`roadmap analyze`** 并对照 **REQUIREMENTS** 全表  
2. **`milestone complete`** 的版本参数勿与 `--help` 混淆（应使用显式版本字符串如 `v1.1`）  

### Cost Observations

- 未单独计量模型用量；与 v1.0 相同可后续补记  

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | 备注 |
|-----------|------|
| v1.0 | 首里程碑，规范 + 测试交付 |
| v1.1 | 最小 Relay + Client + E2E，`go test` 可重复 |

### Cumulative Quality

| Milestone | 测试 | 备注 |
|-----------|------|------|
| v1.0 | `go test ./...` + golden | CI 已接 |
| v1.1 | `go test ./...` + relay/client 集成 | 含 TLS 上负例 |

### Top Lessons（跨里程碑待验证）

1. 需求文件与可追溯表单一事实源（或自动化同步）
