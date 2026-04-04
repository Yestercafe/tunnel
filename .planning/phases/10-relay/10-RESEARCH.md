# Phase 10 — Technical Research

**Phase:** 10 — Relay 数据面路由  
**Gathered:** 2026-04-04  
**Status:** Complete — ready for planning

## Question

如何在 **`pkg/relay`** 已有 **SessionRegistry（peer_id → FrameWriter）** 上实现 **RLY-03**：**JOIN_ACK 之后** 对 **`STREAM_DATA`** 做 **广播（不回送发送者）** 与 **单播（按 dst_peer_id）**，且 **src_peer_id / routing_mode / dst** 与 **`protocol.ValidateRoutingIntent`**、**`DecodeStreamData`** 一致？

## Codebase anchors

| 组件 | 路径 | 说明 |
|------|------|------|
| 当前占位 | `pkg/relay/control.go` — `MsgTypeStreamData` 分支 | Phase 9 返回「phase 10 未实现」；本阶段替换为真实转发 |
| 解码 | `pkg/protocol/streamdata.go` — `DecodeStreamData` | 含 18 字节路由前缀 |
| 路由语义 | `pkg/protocol/routing.go` — `RoutingModeBroadcast` / `Unicast`, `ValidateRoutingIntent` | 广播要求 `dst_peer_id==0` |
| 注册表 | `pkg/relay/registry.go` — `session.peers map[uint64]*Peer` | 需按 **session_id** 选表；JOIN 后连接须绑定 **sessionID** |
| 客户端 | `pkg/client` — `SendStreamData`, `ReadFrame` | 集成测试对端 |

## Recommended approach

1. **`JoinSession` 返回 `(peerID, sessionID, err)`**（或等价地在 **`connState`** 写入 **`sessionID`**），以便数据面查 **`session.peers`**。
2. **入站 `STREAM_DATA`**：`DecodeStreamData` → 校验 **`Prefix.SrcPeerID == st.peerID`**（否则 **`ERR_ROUTING_INVALID`**）→ **`ValidateRoutingIntent`**。
3. **广播**：`RoutingModeBroadcast` → 对 **`sess.peers`** 中 **除 `senderPeerID`** 外每个 **`Write(AppendFrame(...))`**。
4. **单播**：`RoutingModeUnicast` → 仅 **`sess.peers[dst_peer_id]`**（若无则 **`ERR_ROUTING_INVALID`** 或 `SESSION_NOT_FOUND` 策略 — 与 `errors.md` 对齐并在测试中固定）。
5. **`STREAM_OPEN` / `STREAM_CLOSE`**：本里程碑成功标准仅明确要求 **`STREAM_DATA`**；可继续返回 **`PROTOCOL_ERROR`**（文档化）或后续扩展 — 计划中写死一种策略。

## Risks

| Risk | Mitigation |
|------|------------|
| 持锁期间写 TLS 阻塞 | 先 **`Copy`** 目标 `[]*Peer` 或 `[]FrameWriter`，**释放 registry 锁** 再 **`Write`** |
| 与 Client 的 `stream_id` 约定 | 复用 Phase 8 测试与 `docs/client-stream-ids.md` 中的 **stream_id 1/2** |

## Validation Architecture

> Nyquist Dimension 8：本阶段以 **`go test ./pkg/relay/...`** 为主，辅以 **`pkg/client`** 多连接场景。

### Feedback dimensions

| Dimension | Phase 10 |
|-----------|----------|
| 自动化 | `go test` 覆盖广播、单播、负例（错误 `src_peer_id` 等） |
| 手工 | 无强制；CLI 非必须 |

### Test infrastructure

- **Quick：** `go test ./pkg/relay -count=1`
- **Full：** `go test ./... -count=1`

### Sampling

- 每任务后：相关包 `go test`
- Wave 结束：全量 `go test ./pkg/relay/...`

---

## RESEARCH COMPLETE
