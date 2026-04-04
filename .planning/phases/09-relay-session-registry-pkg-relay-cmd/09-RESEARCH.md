# Phase 9 — Technical Research

**Phase:** 9 — Relay 监听与 Session Registry（pkg/relay + cmd）  
**Gathered:** 2026-04-04  
**Status:** Complete — ready for planning

## Question

What must we know to implement **RLY-01** (TCP 监听 + TLS 终止 + 每连接 `ParseFrame`/`ErrNeedMore` 成帧循环) and **RLY-02**（进程内 Session Registry、SESSION_CREATE / SESSION_JOIN、`peer_id` 分配）而不与既有 **`pkg/protocol`**、**`pkg/framing`**、**`pkg/client`** 行为冲突？

## Codebase anchors

| Area | Location | Relevance |
|------|----------|-----------|
| 帧解析 | `pkg/framing/decode.go` — `ParseFrame`, `ErrNeedMore`, `AppendFrame` | Relay 读循环 **MUST** 与之一致 |
| 控制面 | `pkg/protocol/control.go` — CREATE/JOIN/ACK/PROTOCOL_ERROR | Registry 与连接处理直接调用 |
| JOIN 前数据面门禁 | `pkg/protocol/join_gate.go` | JOIN_ACK 前收到 `STREAM_DATA` → 拒绝路径（PROT-02） |
| 客户端参考 | `pkg/client/client.go` | 集成测试对端行为 |
| 测试 harness（非 relay） | `internal/fakepeer/harness.go` | 行为对照；**`pkg/relay` 包名**用于真实 Relay |
| CLI 模式 | `cmd/tunnel/main.go` — `client` 子命令 | 新增 **`relay`** 子命令与 flag 风格对齐 |

## Recommended architecture

1. **`pkg/relay.Server`**：`net.Listen("tcp", addr)` + `tls.NewListener`；每接受连接启动 **goroutine**，持 **`[]byte` 读缓冲**，循环 **`framing.ParseFrame`**；写回 **`tls.Conn`** 使用 **`framing.AppendFrame`**。
2. **错误路径**：`ErrFrameTooLarge` / `ErrProtoVersion` → **`protocol.EncodeProtocolError`** + `AppendFrame`；与 `docs/spec/v1/errors.md` 中 `ErrCode` 对齐。
3. **`SessionRegistry`**（进程内）：`session_id`（UUID 字符串）→ session 元数据 + `peer_id` 分配器 + peer 表（`peer_id` → 可写 `*tls.Conn` 或抽象 **`FrameWriter`**）。**CREATE** 生成新 `session_id`（`github.com/google/uuid` 等与协议校验一致）、**invite_code**（与 `protocol` 校验规则一致）。**JOIN** 按 `join_by` + credential 查表。
4. **Phase 9 边界**：**不**实现 RLY-03（JOIN 后 `STREAM_DATA` 广播/单播转发）；对 **`MsgTypeStreamData`** 在 JOIN_ACK 前或 Phase 9 占位策略：**`PROTOCOL_ERROR`** 或文档化丢弃 — 与 **`JoinGateAllowsBusinessDataPlane`** 一致。

## Risks

| Risk | Mitigation |
|------|------------|
| TLS 证书配置拖累测试 | 测试复用 **`internal/fakepeer`** 的 **`LocalhostTLSConfig`** 模式或 PEM `testdata/` |
| Registry 与连接生命周期竞态 | `sync.Mutex` 每 session 或全局 registry lock；连接关闭时注销 peer |
| 与 fakepeer 行为漂移 | Phase 9 验收用 **`pkg/client`** 对 **`pkg/relay` Server** 跑通 CREATE/JOIN（09-02） |

## Validation Architecture

> Nyquist / Dimension 8：本阶段验证策略与自动化命令的**单一权威**说明。

### Feedback dimensions

| Dimension | How Phase 9 satisfies |
|-----------|----------------------|
| 1–7 | `go test` 覆盖成帧、控制面 ACK、registry 不变量；无 UI |
| 8 | 本文件 + `09-VALIDATION.md` 定义每任务后的 **`go test`** 采样命令 |

### Test infrastructure

- **Framework：** Go `testing` + 标准库 `crypto/tls`、`net`
- **Quick（每任务后）：** `go test ./pkg/relay/... -count=1`（或当前包路径）
- **Full（每 wave / PR 前）：** `go test ./... -count=1`
- **Integration：** `09-02` 使用 **`pkg/client`** 对本地 **`relay.Server`** 端口跑 CREATE+JOIN

### Instrumentation

- 断言 **`protocol.DecodeSessionCreateAck` / `DecodeSessionJoinAck`** 与 **`framing.ParseFrame`** 输出一致
- 对 **PROTOCOL_ERROR** 响应使用 **`protocol.DecodeProtocolError`** 在负例测试中解码（可选）

### Sampling policy

- 每个 **task commit** 后：至少运行 **相关包** `go test`（计划中 `<automated>`）
- Wave 1 结束：`go test ./pkg/relay/...`
- Wave 2 结束：`go test ./...`

---

## RESEARCH COMPLETE

Plans may assume **`docs/spec/v1/`** 与 **`pkg/protocol`** 为控制面真值来源；**`pkg/framing`** 为帧边界真值来源。
