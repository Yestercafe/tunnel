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

## Cross-Milestone Trends

### Process Evolution

| Milestone | 备注 |
|-----------|------|
| v1.0 | 首里程碑，规范 + 测试交付 |

### Cumulative Quality

| Milestone | 测试 | 备注 |
|-----------|------|------|
| v1.0 | `go test ./...` + golden | CI 已接 |

### Top Lessons（跨里程碑待验证）

1. 需求文件与可追溯表单一事实源（或自动化同步）
