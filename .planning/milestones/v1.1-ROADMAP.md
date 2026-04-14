# Roadmap: Tunnel（传输隧道）

## Milestones

- ✅ **v1.0 — v1 协议规范与一致性测试** — Phases 1–6（[归档与完整阶段说明](milestones/v1.0-ROADMAP.md)），交付日 **2026-04-04**
- 🚧 **v1.1 — 最小 Relay 与 Client** — Phases **7–11**（本文件下文；首阶段目录 **`07-*`**）

---

## v1.1 — 最小 Relay 与 Client（进行中）

**Milestone goal：** 在 **TCP+TLS** 上交付与 `docs/spec/v1/` 一致的 **Go 最小 Relay** 与 **Client**，跑通创建 session、加入、同 session 内广播与单播；含可重复的自动化测试与负例门禁。

**依赖顺序（实现建议）：** 共享协议层 → Client 控制/数据路径 → Relay 监听与 Registry → Relay 数据面 → E2E 测试。

---

## Phases

- [ ] **Phase 7: 协议载荷层（pkg/protocol）** — 控制面/数据面编解码与 JOIN 前数据面门禁（PROT-01、PROT-02）
- [x] **Phase 8: Client（pkg/client + cmd）** — TCP+TLS、SESSION_CREATE/JOIN、JOIN 后 STREAM_DATA 广播与单播（CLNT-01..03） (completed 2026-04-04)
- [x] **Phase 9: Relay 监听与 Session Registry（pkg/relay + cmd）** — TCP+TLS、成帧循环、CREATE/JOIN 与 peer_id（RLY-01、RLY-02） (completed 2026-04-04)
- [x] **Phase 10: Relay 数据面路由** — JOIN 后广播（不回送）与单播转发、非法路径行为可测（RLY-03） (completed 2026-04-04)
- [x] **Phase 11: E2E 验证与负例** — 双 Client 同 session 自动化测试；JOIN 前数据面或非法路由门禁（E2E-01、E2E-02） (completed 2026-04-14)

---

## Phase Details

### Phase 7: 协议载荷层（pkg/protocol）

**Goal**：Relay 与 Client 共用、与 v1 规范字段级一致的控制面/数据面语义层；**JOIN_ACK 之前**不得将带数据面路由的帧视为合法业务路径。

**Depends on**：Phase 6（v1.0 成帧与规范基线）

**Requirements**：PROT-01, PROT-02

**Success Criteria**（what must be TRUE）:

1. 工程师可对控制面/数据面 payload、`msg_type`、路由前缀及 **`PROTOCOL_ERROR`** 进行编解码，且与 v1 规范字段对齐；与 `pkg/framing` 边界清晰（帧级 vs 会话语义）。
2. **JOIN_ACK 前**，共享判定逻辑将「带数据面路由」的帧视为非合法业务路径；Client/Relay 可一致实现或复用同一门禁 helper。

**Plans**：2 plans

Plans:

- [ ] `07-01-PLAN.md` — `pkg/protocol` 骨架、PROT-01 控制面/数据面编解码、`PROTOCOL_ERROR`、STREAM_* 视图与 golden
- [ ] `07-02-PLAN.md` — PROT-02 `join_gate` 门禁、JoinGate 测试与包文档、全量 `go test ./...`

### Phase 8: Client（pkg/client + cmd）

**Goal**：Client 在 **TCP+TLS** 上完成会话创建与加入，并在 **JOIN_ACK 之后**收发 **`STREAM_DATA`**（含广播与单播各至少一条可重复路径）。

**Depends on**：Phase 7

**Requirements**：CLNT-01, CLNT-02, CLNT-03

**Success Criteria**（what must be TRUE）:

1. Client 通过 **TCP+TLS** 连接 Relay，完成 **SESSION_CREATE**，获得与规范一致的 `session_id` 与 `invite_code`。
2. Client 可凭 `session_id` 或邀请码 **SESSION_JOIN**，在 **SESSION_JOIN_ACK** 后获得非 0 的 **`peer_id`**。
3. **JOIN_ACK 之后**，Client 可发送并接收带路由前缀的 **`STREAM_DATA`**；至少验证 **广播**与**单播**各一条可重复路径（`stream_id` 等策略已文档化）。

**Plans**：3 plans

Plans:

- [ ] `08-01-PLAN.md` — `internal/fakepeer` TLS harness（SESSION_CREATE/JOIN、STREAM_DATA 路由；非 relay 包名）
- [ ] `08-02-PLAN.md` — `pkg/client`（Dial、CreateSession、JoinSession、STREAM_DATA、JoinGate、CLNT-01..03 测试 + `docs/client-stream-ids.md`）
- [ ] `08-03-PLAN.md` — `cmd/tunnel`（`--addr`、`--insecure-skip-verify`、client 子命令冒烟）

### Phase 9: Relay 监听与 Session Registry（pkg/relay + cmd）

**Goal**：Relay **监听 TCP**、**TLS 终止**、每连接成帧循环；进程内 **Session Registry** 处理 **SESSION_CREATE** / **SESSION_JOIN** 并分配 **`peer_id`**。

**Depends on**：Phase 7（Phase 8 可与本阶段并行开发思想验证，但集成依赖上通常先有协议层稳定）

**Requirements**：RLY-01, RLY-02

**Success Criteria**（what must be TRUE）:

1. Relay **监听 TCP** 并以 **TLS** 终止连接；每条连接维护读缓冲与成帧循环（`ParseFrame` / `ErrNeedMore`），行为与 `pkg/framing` 一致。
2. Relay 维护进程内 **Session Registry**（session ↔ peer ↔ 可写连接），正确处理 **SESSION_CREATE** / **SESSION_JOIN**，并分配 **`peer_id`**。

**Plans**：2 plans

Plans:

- [ ] `09-01-PLAN.md` — `pkg/relay`：TCP+TLS 监听、每连接读缓冲与 `ParseFrame`/`ErrNeedMore` 成帧循环、`cmd/tunnel relay`（RLY-01）
- [ ] `09-02-PLAN.md` — Session Registry、SESSION_CREATE / SESSION_JOIN、`peer_id` 分配、`pkg/client` 集成测试（RLY-02）

### Phase 10: Relay 数据面路由

**Goal**：在 **JOIN_ACK 之后**，对 **STREAM_DATA** 执行广播（不回送发送者）与单播（按 `dst_peer_id`）；非法或未 JOIN 行为与 `errors.md` 一致且可测。

**Depends on**：Phase 9

**Requirements**：RLY-03

**Success Criteria**（what must be TRUE）:

1. **JOIN_ACK 之后**，Relay 对 **STREAM_DATA** 执行**广播**，且**不**回送发送者。
2. Relay 对 **STREAM_DATA** 执行**单播**：按 `dst_peer_id` 查表投递。
3. 对非法路由或尚未 JOIN 的连接，行为与 `errors.md` 及实现注释一致，且测试可断言（含 `PROTOCOL_ERROR`、丢弃或断连等选定策略）。

**Plans**：2 plans

Plans:

- [ ] `10-01-PLAN.md` — `JoinSession` 返回 sessionID、`DeliverStreamData`、`control.go` 解码与广播/单播（RLY-03）
- [ ] `10-02-PLAN.md` — `relay_test` 广播/单播/负例（RLY-03）

### Phase 11: E2E 验证与负例

**Goal**：自动化（或 CI 可运行脚本）证明两 Client 同 Relay、同 session 的广播与单播；并覆盖 JOIN 前数据面或非法路由负例。

**Depends on**：Phase 8, Phase 10（端到端需 Client + Relay 数据面齐备）

**Requirements**：E2E-01, E2E-02

**Success Criteria**（what must be TRUE）:

1. **自动化测试**（或同等 CI 脚本）：至少**两个 Client** 连接同一 Relay、**同一 session**，可验证 **广播**与**单播**各至少一条用例。
2. **负例或门禁**：至少覆盖 **JOIN_ACK 前发送数据面帧** 与 **非法/未知路由** 之一；期望 **`PROTOCOL_ERROR`**、丢弃或断连与实现策略一致，并在测试中断言。

**Plans**：2 plans

Plans:

- [x] `11-01-PLAN.md` — 追溯与文档（E2E-01、E2E-02）
- [x] `11-02-PLAN.md` — `TestRelay_StreamData_BeforeJoinAck` + `UnderlyingTLSConn`（E2E-02）

---

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 7. 协议载荷层 | 2/2 | Planned | - |
| 8. Client | 0/3 | Complete    | 2026-04-04 |
| 9. Relay 监听与 Registry | 2/2 | Complete   | 2026-04-04 |
| 10. Relay 数据面路由 | 0/TBD | Complete    | 2026-04-04 |
| 11. E2E 验证与负例 | 2/2 | Complete    | 2026-04-14 |

---

*v1.0 历史阶段 1–6 见 [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)。*
