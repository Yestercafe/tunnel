# Phase 11: E2E 验证与负例 - Research

**Researched:** 2026-04-05  
**Domain:** Go 集成/E2E 测试（`testing`、进程内 Relay、`crypto/tls`、v1 成帧）  
**Confidence:** HIGH（结论主要来自仓库内现有测试与实现代码）

<user_constraints>

## User Constraints（无 discuss 阶段 CONTEXT.md）

本阶段目录下不存在 `*-CONTEXT.md`，故无「Decisions / Claude's Discretion / Deferred Ideas」逐字拷贝。以下约束来自 `.planning/REQUIREMENTS.md`、`.planning/ROADMAP.md` Phase 11 与 `.cursor/rules/gsd-project.md`，**规划阶段须遵守**。

### 来自 REQUIREMENTS / Roadmap（锁定范围）

- **E2E-01**：自动化（或同等 CI 脚本）：至少两个 Client、同一 Relay、同一 session；**广播**与**单播**各至少一条用例。
- **E2E-02**：负例或门禁：至少覆盖 **JOIN_ACK 前发送数据面帧** 与 **非法/未知路由** 之一；期望 **`PROTOCOL_ERROR`**、丢弃或断连须与实现一致并在测试中断言。
- **v1 承载**：**TCP+TLS**（不采用 WebSocket 作为 v1 主路径）。
- **依赖**：Phase 8（Client）与 Phase 10（Relay 数据面）已交付后方可认为本阶段端到端前提满足。

### 来自 .cursor/rules/gsd-project.md（项目级）

- 实现默认 **Go**；传输为 **TLS 上的自定义成帧**。
- 测试运行以 **`go test`** 为主（与 `research/STACK.md` / CI 一致）。

</user_constraints>

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| E2E-01 | 双 Client、同 Relay、同 session；广播 + 单播各 ≥1 条自动化用例 | 复用/提炼 `pkg/relay/relay_test.go` 中 `TestRelay_StreamData_Broadcast`、`TestRelay_StreamData_Unicast` 模式；TLS 用 `fakepeer.LocalhostTLSConfig` |
| E2E-02 | JOIN 前数据面 **或** 非法路由负例；`PROTOCOL_ERROR` 等断言 | 非法路由：`TestRelay_StreamData_UnicastMissingDst`；JOIN 前数据面：需对 **真实 Relay** 发原始 `STREAM_DATA`（见下文），断言 `pkg/relay/control.go` 返回的 `PROTOCOL_ERROR` |

</phase_requirements>

## Summary

本阶段目标是在 **CI 可运行** 的前提下，用自动化测试证明：**两个** `pkg/client.Client` 连接**同一** `pkg/relay.Server`、**同一** `session_id`，并完成 **广播**与**单播**各至少一条断言；负例需覆盖 **JOIN_ACK 之前的数据面帧** 与 **非法/未知路由** 中的至少一类，并对 **`PROTOCOL_ERROR`**（或实现选定的等价行为）做稳定断言。

仓库内**已存在**高度接近 E2E-01/部分 E2E-02 的实现：`pkg/relay/relay_test.go` 使用进程内 `relay.Server` + `internal/fakepeer.LocalhostTLSConfig` + `client.Dial`，完成双 peer 建 session、JOIN、广播与单播；负例 **非法单播目标** 已由 `TestRelay_StreamData_UnicastMissingDst` 覆盖（期望 `PROTOCOL_ERROR` + `ErrCodeRoutingInvalid`）。Phase 11 的增量工作主要是：**明确「E2E」归属与可追溯性**（是否单独 `e2e` 包/文件/构建标签）、补齐 **针对真实 Relay** 的 **JOIN 前数据面**负例（`pkg/client` 的 `SendStreamData` 在本地门禁即返回 `ErrNotJoined`，不会发到线上，需 **原始帧写入** 或测试辅助函数），以及将需求 E2E-01/E2E-02 与测试函数在文档或 `REQUIREMENTS` 追溯表中显式对齐。

**Primary recommendation：** 以 **进程内** `relay.Server` 为主 harness（与现有 `relay_test` 一致），**不**强制 `cmd/tunnel` 子进程；负例优先复用 `relay` 侧已有路由错误路径，并新增一条 **绕过 Client 发送 API** 的「JOIN 前 `STREAM_DATA`」用例以贴合 E2E-02 字面要求。

## Project Constraints（from .cursor/rules/）

| 来源 | 指令摘要 |
|------|----------|
| `gsd-project.md` | Go 实现；先规范与测试再扩展实现；v1 不用 WebSocket；安全边界以 TLS 为主 |
| `gsd-project.md` / STACK | `go test ./...` 为一致性测试入口；可选用 `golangci-lint` |

## Standard Stack

### Core

| 组件 | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Go | `go 1.22`（`go.mod`）；CI 固定 `1.22`（`.github/workflows/go.yml`） | 测试与实现 | 项目基线 |
| `crypto/tls` + `net` | 标准库 | 测试中的 TLS 服务端/客户端 | 与生产路径一致 |
| `internal/fakepeer` | 仓库内 | 测试用本地 TLS 证书与（Phase 8）假 peer | `LocalhostTLSConfig` 被 `relay_test` / `client_test` 共用 |

**说明：** 仓库当前 **未** 使用 `testify`（`grep` 无匹配）；测试风格为 `testing` + `t.Fatal`/`t.Fatalf`，与现有文件一致即可。

**Version verification：** `go.mod` 声明 `go 1.22`；未引入额外版本化测试库。

### Supporting

| 组件 | Purpose | When to Use |
|------|---------|-------------|
| `pkg/framing` | 组帧/解析 | 若需发送「绕过 Client API」的原始帧 |
| `pkg/protocol` | `EncodeStreamData`、`DecodeProtocolError`、`MsgTypeStreamData` 等 | 构造/解析负例期望 |

### Alternatives Considered

| 方案 | 说明 | Tradeoff |
|------|------|----------|
| 进程内 `relay.Server`（当前） | `ListenAddr: "127.0.0.1:0"`，`Serve(ctx)` 于 goroutine | 快速、无二进制依赖、CI 友好 |
| 子进程 `cmd/tunnel relay` | 黑盒测 CLI + 真实进程 | 更慢、需 PEM 证书路径（`--cert`/`--key`）；`runRelay` 未设置 `Registry`，`Listen()` 在 `Registry == nil` 时注入 `NewSessionRegistry()`（见 `pkg/relay/server.go`），与进程内测试等价 |

**Installation：** 仅 Go 工具链；依赖模块首次构建需网络（如 `github.com/google/uuid`）。

## Architecture Patterns

### 推荐：进程内 E2E（与现有一致）

**What：** 测试内构造 `relay.Server{ListenAddr, TLSConfig, Registry}` → `Listen()` → `go Serve(ctx)` → 两个 `client.Dial` → `CreateSession` / `JoinSession` → `SendStreamData` / `ReadFrame`。

**When：** 默认所有 Phase 11 自动化；与 `pkg/relay/relay_test.go` 相同。

**参考结构：**

```
pkg/relay/relay_test.go   # 已含双 client、广播、单播、非法单播目标
pkg/client/client_test.go # fakepeer 上双 client；含 SendStreamData 未 JOIN（ErrNotJoined）
internal/fakepeer/        # LocalhostTLSConfig、Harness（非生产 Relay）
```

### Pattern：双 Client 同 Session

1. `c1.CreateSession` → `sid`（与 `invite`）。  
2. `c1.JoinSession(ctx, true, sid)` → `p1`。  
3. `c2.Dial`（**新 TCP+TLS 连接**）→ `c2.JoinSession(ctx, true, sid)` → `p2`，且 `p1 != p2`。  
4. 广播：`c1.SendStreamData(..., RoutingModeBroadcast, 0, streamID, ...)`，`c2.ReadFrame` + `protocol.DecodeStreamData`。  
5. 单播：需要目标 `peer_id`（使用 `c2` 的 `JoinSession` 返回值或 `client` 在 ACK 后设置的 `PeerID()`）；`readUntilStreamData` 模式见 `pkg/client/client_test.go` 的 `readUntilStreamData`。

### Pattern：TLS 与端口

- **证书：** `fakepeer.LocalhostTLSConfig(t)`（`internal/fakepeer/tlscert.go`）：ECDSA P-256，SAN 含 `localhost` 与 `127.0.0.1`。  
- **客户端：** `tls.Config{RootCAs: pool, ServerName: "localhost", MinVersion: tls.VersionTLS12}`（见 `relay_test.go`）。`pkg/client` 在 `ServerName` 为空且 host 非 IP 时会自动设置（`client.go`）。  
- **端口：** `"127.0.0.1:0"` 动态分配；`srv.Addr().String()` 传给 `Dial`。

### Pattern：定时与竞态

- 使用 `context.WithTimeout`（现有测试为 5–15s 级）包裹 `SendStreamData`/`ReadFrame`。  
- `Serve` 在单独 goroutine；`defer srv.Close()` 与 `cancel()` 清理。

### Anti-Patterns to Avoid

- **仅用 `client.SendStreamData` 测「JOIN 前数据面」：** `pkg/client/client.go` 在发送路径调用 `JoinGateAllowsBusinessDataPlane`，未 JOIN 时直接 **`return ErrNotJoined`**，帧**不会**到达 Relay，无法验证服务端 `PROTOCOL_ERROR` 路径。负例须 **`writeFrame` 级**发送或共享测试辅助函数。  
- **混淆 fakepeer 与 Relay：** `internal/fakepeer/harness.go` 对 JOIN 前数据面返回的 `EncodeProtocolError` 与 `pkg/relay/control.go` 的 **err_code/reason 字符串** 可能不一致；E2E-02 若要求「真实 Relay」，应以 **`relay.Server`** 为准。

## Runtime State Inventory

> 本阶段为**验证/测试组织**与可选新增 `_test.go`，**非**重命名、数据迁移或运行时配置变更。

| Category | Items Found | Action Required |
|----------|-------------|-----------------|
| Stored data | 无（测试不写 DB/无持久化 session） | — |
| Live service config | 无 | — |
| OS-registered state | 无 | — |
| Secrets/env vars | 无新增密钥需求；测试仅用 `fakepeer` 内存证书 | — |
| Build artifacts | 无；`go test` 不产生需清理的安装态 | — |

**结论：** Phase 11 交付不涉及仓库外运行时状态同步；若未来增加「子进程 + 临时 PEM 文件」脚本，仅需在测试中 `t.TempDir()` 内生成并删除。

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 测试用 TLS 证书 | 每测试手写 PEM | `fakepeer.LocalhostTLSConfig(t)` | 已验证与 `127.0.0.1`/`localhost` 匹配 |
| 帧边界 | 手动拼字节 | `framing.AppendFrame` + `framing.ParseFrame` | 与 `pkg/framing` 一致 |
| JOIN 门禁语义 | 各测各写 | `protocol.JoinGateAllowsBusinessDataPlane` / 读 `pkg/relay/control.go` | 与 PROT-02 / RLY-03 对齐 |

## Common Pitfalls

### Pitfall 1：负例「JOIN 前数据面」测错层

**现象：** 断言了 `ErrNotJoined`，但需求要的是 Relay/线上 **`PROTOCOL_ERROR`**。  
**原因：** Client 发送侧门禁先于写 socket。  
**避免：** 在 `relay_test`（或等价包）中持有 `tls.Conn` 或提供 **仅测试可见** 的 `WriteRawFrame`，发送合法编码的 `STREAM_DATA`，再 `ReadFrame` 断言 `MsgTypeProtocolError`（见 `control.go` 中 `MsgTypeStreamData` + `!ok` 分支，`ErrCodeRoutingInvalid` + `"ERR_ROUTING_INVALID"`）。

### Pitfall 2：ServerName 与证书不一致

**现象：** TLS 握手失败。  
**避免：** 与 `relay_test` 一致使用 `ServerName: "localhost"` 且连 `127.0.0.1`，或统一用 `127.0.0.1` 作 ServerName（需与 `tlscert.go` SAN 一致）。

### Pitfall 3：单播断言用错 peer

**现象：** 收不到包或路由错误。  
**避免：** 单播 `dst_peer_id` 必须为对端 JOIN 后分配的 ID（`TestRelay_StreamData_Unicast` 使用 `p2`）。

## Code Examples（摘自仓库）

### 进程内 Relay + 双 JOIN（节选）

```go
// 来源：pkg/relay/relay_test.go — TestRelay_ClientCreateJoin
srvCfg, pool := fakepeer.LocalhostTLSConfig(t)
srv := &relay.Server{ListenAddr: "127.0.0.1:0", TLSConfig: srvCfg, Registry: relay.NewSessionRegistry()}
// Listen / Serve ...
tlsClient := &tls.Config{RootCAs: pool, ServerName: "localhost", MinVersion: tls.VersionTLS12}
// c1 CreateSession + Join; c2 Dial + JoinSession(..., sid)
```

### Relay 对「未 JOIN 的 STREAM_DATA」响应（实现事实）

```91:106:pkg/relay/control.go
	case protocol.MsgTypeStreamData:
		ok, err := protocol.JoinGateAllowsBusinessDataPlane(st.joined, f.Payload)
		// ...
		if !ok {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "ERR_ROUTING_INVALID")
			// ...
			return writeRawFrame(conn, p)
		}
		return s.dispatchStreamData(conn, reg, st, f)
```

### 非法单播目标（已有负例）

```213:269:pkg/relay/relay_test.go
func TestRelay_StreamData_UnicastMissingDst(t *testing.T) {
	// ... Join 后 SendStreamData(unicast, dst=999) ...
	code, _, err := protocol.DecodeProtocolError(f.Payload)
	if code != framing.ErrCodeRoutingInvalid { /* ... */ }
}
```

### JOIN 前数据面（待补；须绕过 `Client.SendStreamData`）

思路：在 `relay_test` 中 `Dial` 后仅 `CreateSession`，**不**调用 `JoinSession`；使用与 `pkg/client` 相同的成帧方式写出 **一帧** `STREAM_DATA`（`protocol.EncodeStreamData` + `framing.AppendFrame`）。因 `Client` 未导出底层 `writeFrame`，可复制 `client.writeFrame` 逻辑到测试辅助函数，或使用 `tls.Dial` + 手写 `Write`。随后用 `protocol.DecodeProtocolError` 读回一帧，期望 `MsgTypeProtocolError` 且 `err_code == framing.ErrCodeRoutingInvalid`（与 `pkg/relay/control.go` 一致）。

**勿用 `internal/fakepeer` 断言 Relay 的 err_code：** fakepeer 在 JOIN 前 `STREAM_DATA` 上返回 `ErrCodeJoinDenied`（见 `internal/fakepeer/harness.go` `MsgTypeStreamData` 分支），与生产 Relay 的 `ErrCodeRoutingInvalid` **不一致**；E2E-02 若以真实 Relay 为准，必须在 `relay.Server` 上测。

## State of the Art

| 旧认知 | 当前仓库事实 | 说明 |
|--------|----------------|------|
| E2E 必须起 `cmd/tunnel` | 进程内 `relay.Server` 已覆盖数据面 | CI 与 `relay_test` 一致即可满足「自动化」 |
| 负例仅靠 `pkg/client` | Client 侧 `ErrNotJoined` 与 Relay 侧 `PROTOCOL_ERROR` 是两条路径 | E2E-02 至少一条须落到 **Relay 可观测行为**（若选 JOIN 前数据面） |

## Open Questions

1. **是否将现有 `relay_test` 重命名/标注为 E2E？**  
   - 已知：行为已覆盖 E2E-01 核心。  
   - 未定：产品/流程是否要求独立目录名 `e2e` 或 `//go:build integration`。  
   - 建议：Phase 计划阶段与「可追溯表」对齐即可；最小改动为增加注释与 REQ 映射。

2. **JOIN 前负例的 err_code 是否必须与 `errors.md` 中某条逐字一致？**  
   - 已知：Relay 使用 `ErrCodeRoutingInvalid`（`0x0005`）与固定 reason 串。  
   - 建议：测试断言 **msg_type + err_code**；reason 字符串与 `control.go` 绑定，若规范后续收紧再 golden。

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | `go test`, CI | ✓（本机有 `go`） | 本机 1.26.x；模块声明 1.22；CI 1.22 | 以 `go.mod` / CI 为准 |
| 网络（module download） | 首次 `go test` 拉取模块 | 视环境 | — | CI 通常有；离线需 vendor 或缓存 |

**无外部 DB/服务依赖**；E2E 为纯进程内网络监听。

**Step 2.6 说明：** 若 Phase 仅新增 `_test.go` 与标准 `go test`，无额外 CLI 依赖。

## Validation Architecture

> 对应 `.planning/config.json` 中 `workflow.nyquist_validation: true`（未关闭），以下为 Nyquist / 验证闭环建议。

### Test Framework

| Property | Value |
|----------|--------|
| Framework | Go `testing`（stdlib） |
| Config file | 无独立 `pytest.ini` 类文件；入口为模块根 `go.mod` |
| Quick run command | `go test ./pkg/relay/... -count=1` 或单测 `-run TestRelay_StreamData` |
| Full suite command | `go test ./... -count=1`（与 `.github/workflows/go.yml` 一致） |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| E2E-01 | 双 Client、同 session、广播 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_Broadcast -count=1` | ✅ `pkg/relay/relay_test.go` |
| E2E-01 | 双 Client、同 session、单播 | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_Unicast -count=1` | ✅ 同上 |
| E2E-02 | 非法单播目标 → `PROTOCOL_ERROR` | 集成 | `go test ./pkg/relay/... -run TestRelay_StreamData_UnicastMissingDst -count=1` | ✅ 同上 |
| E2E-02 | JOIN 前数据面 → Relay `PROTOCOL_ERROR` | 集成 | 待新增：`-run TestRelay_...PreJoin`（命名示例） | ❌ Wave 0 缺口 |

### Sampling Rate（建议）

- **Per task / wave merge：** `go test ./... -count=1`  
- **Phase gate：** 全绿后进入 `/gsd-verify-work`；CI 与本地命令对齐（`.github/workflows/go.yml`）。
- **可选加强：** `go test ./... -race -count=1`（若新增并发相关逻辑；当前集成测试以同步顺序为主）。

### Wave 0 Gaps

- [ ] **新增** 针对 `relay.Server` 的 **JOIN 前发送 `STREAM_DATA`** 测试：构造 `protocol.EncodeStreamData` + `framing.AppendFrame`，经 **未调用 `JoinSession` 的连接** 写出；`ReadFrame` 断言 `PROTOCOL_ERROR`（与 `pkg/relay/control.go` 一致）。  
- [ ] （可选）在 `REQUIREMENTS.md` Traceability 或本阶段 `VERIFICATION.md` 中将上述用例 **显式映射到 E2E-01/E2E-02**。  
- [ ] 若保留仅 Client 侧 `TestClient_SendStreamData_NotJoined`：在文档中标注其 **不满足**「Relay 收到数据面」的 E2E-02 子项，避免验收歧义。

**若仅将「已有 relay 测试」声明为 Phase 11 交付：** 则缺口收缩为 **文档/追溯 + 可选一条 JOIN 前原始帧**。

## Sources

### Primary（HIGH）

- `pkg/relay/relay_test.go` — 双 client、广播、单播、非法路由  
- `pkg/relay/control.go` — JOIN 前 `STREAM_DATA` 与 `PROTOCOL_ERROR`  
- `pkg/client/client.go` — `SendStreamData` 门禁与 `ErrNotJoined`  
- `internal/fakepeer/tlscert.go`、`internal/fakepeer/harness.go` — TLS 与假 peer 行为  
- `.planning/REQUIREMENTS.md` — E2E-01、E2E-02  
- `.github/workflows/go.yml` — CI 测试命令  

### Secondary（MEDIUM）

- `docs/spec/v1/errors.md` — `ERR_ROUTING_INVALID` 语义  
- `cmd/tunnel/main.go`、`cmd/tunnel/relay.go` — 子进程 E2E 可选路径  

## Metadata

**Confidence breakdown：**

- Standard stack: **HIGH** — 全部来自 `go.mod` 与现有测试导入。  
- Architecture: **HIGH** — 与 `relay_test.go` / `client_test.go` 一致。  
- Pitfalls: **HIGH** — 「Client 门禁 vs Relay 负例」由源码路径直接推出。  

**Research date:** 2026-04-05  
**Valid until:** ~30 天（测试布局若重构需重读）

## RESEARCH COMPLETE
