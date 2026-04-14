# Milestones

## v1.1 最小 Relay 与 Client (Shipped: 2026-04-14)

**Phases completed:** 5 phases, 11 plans, 24 tasks

**Key accomplishments:**

- 交付 `pkg/protocol` PROT-01
- 实现 PROT-02 `join_gate`
- Minimal TLS + v1 framing fake (`internal/fakepeer`) enables `pkg/client` integration tests without Docker or production Relay.
- `pkg/client` delivers CLNT-01..03 with `go test` against `internal/fakepeer`, plus stream_id documentation.
- `cmd/tunnel` exposes `client create` and `client join` for manual/script smoke against a relay or fake.
- Delivered `pkg/relay` TLS TCP listener, per-connection framing loop with PROTOCOL_ERROR on frame errors, and `tunnel relay --listen --cert --key`.
- Delivered process-local Session Registry, SESSION_CREATE / SESSION_JOIN handling with non-zero peer_id, JOIN 前 STREAM_DATA 拒绝，以及 `pkg/client` 对真实 Relay 的集成测试。
- Extended `JoinSession` with `sessionID`, added `DeliverStreamData` (broadcast excludes sender; unicast by `dst_peer_id`), and replaced the Phase 9 STREAM_DATA placeholder with `DecodeStreamData` + `ValidateRoutingIntent` routing.
- Added `TestRelay_StreamData_Broadcast`, `TestRelay_StreamData_Unicast`, and `TestRelay_StreamData_UnicastMissingDst` (unicast to absent peer → `PROTOCOL_ERROR` `ERR_ROUTING_INVALID`).
- 将 E2E-01/E2E-02 与已运行的 `go test` 路径、路线图与验证表对齐，完成可追溯性交付。
- 在真实 `relay.Server` 上补齐 JOIN_ACK 前 `STREAM_DATA` 负例，与 `control.go` 门禁一致（`ErrCodeRoutingInvalid`）。

---

## v1.0 v1 协议规范与一致性测试 (Shipped: 2026-04-04)

**Delivered:** v1 二进制协议规范（帧、会话、路由/流、应用信封、状态机/错误/安全假设）与 Go 一致性测试 + `testdata` golden，CI 可重复运行 `go test ./...`。

**Phases completed:** 6 phases, 16 plans, 17 tasks

**Stats:** ~477 LOC Go（当前仓库）；时间线 2026-03-29（首交）→ 2026-04-04（里程碑归档）。

**Key accomplishments:**

- 建立了 v1 规范目录，并成文固定 10 字节帧头 + payload 布局（FRAME-01）。
- 定义 version（偏移 4–5）与 capability（偏移 6–9）的语义、v1 仅支持 0x0001、未知 capability 位必须忽略；含非零示例帧。
- 撰写 TLS 字节流上的成帧状态机（半包/粘包）与错误占位；新增 `pkg/framing` 参考实现。
- v1 控制面定义了 SESSION_CREATE/JOIN 四类 opcode、session_id（UUID 字符串）与 invite_code（Base32），并给出 Relay 最小成员表与 ASCII 序列步骤。
- 定义 peer_id 为会话内唯一的 uint64（大端、0 保留），由 Relay 在 JOIN 成功后分配，并与 SESSION_JOIN_ACK 字段一致。
- 约定可选 join_token 在 JOIN 体中的二进制布局与长度上限，并固定 ERR_JOIN_DENIED / ERR_SESSION_NOT_FOUND 占位名以待 Phase 5。
- v1 数据面广播路由前缀（`routing_mode`/`src`/`dst`）与 Relay 不回送规则已写入 `routing-modes.md`，README 可发现。
- 在同文件内扩展 ROUTE-02：单播枚举、dst 规则、Relay 行为与完整帧示例；README 索引同步。
- 新增 `streams-lifecycle.md`：OPEN/DATA/CLOSE、`stream_id`、FIN/CLOSE 优先级、流内/流间顺序与多流文字用例；README 索引更新。
- UTF-8 JSON 应用信封（APP-01）、`HAS_APP_ENVELOPE` 位语义，以及 `pkg/appenvelope` 与规范一致的 `application_data` 切分
- 在 `app-envelope.md` 增补可抄写的端到端示例 A/B（偏移表与十六进制前缀），覆盖 JSON 请求/响应与 Copilot 关联 id 场景
- 交付连接级与 session 级状态终稿文档
- 建立 ERR-01 唯一权威 `errors.md`
- 新增 `security-assumptions.md`
- 在仓库根建立 `testdata/` 分层与命名说明，并加入最小 GitHub Actions，使 PR 上可重复运行 `go test ./... -count=1`（Go 1.22）。
- 在 06-01 约定下补齐成帧与信封切分的 testdata 驱动用例，并加入 routing-modes 广播整帧 golden，`go test ./... -count=1` 全绿。

---
