# Phase 8: Client（pkg/client + cmd）- Research

**Researched:** 2026-04-04  
**Domain:** Go client library + CLI over TCP+TLS, v1 framing + protocol (CLNT-01..03)  
**Confidence:** HIGH (codebase + specs inspected); MEDIUM on exact `Client` method names (planner discretion per D-01..D-12)

<user_constraints>
## User Constraints (from 08-CONTEXT.md)

### Locked Decisions
- **D-01:** Full E2E (two peers, same session, broadcast + unicast, CI) is **Phase 11**; Phase 8 must not block Phase 9–11; public API + tests must stay stable for Relay/E2E.
- **D-02:** CLNT-01..03 proven via **minimal in-repo fake/harness** (not production Relay), **`go test`**, same path as **TLS + framing + `pkg/protocol`**, aligned with **`docs/spec/v1/`**.
- **D-03:** Primary verification = **`go test ./...`**; **no Docker**; fake implements minimal subset (SESSION_CREATE, SESSION_JOIN, JOIN_ACK, STREAM_DATA broadcast/unicast fields).
- **D-04:** Fake placement/naming by plan; must be readable; **must not** use package name `relay` for fake.
- **D-05:** Single binary, one `main`; client subcommands for smoke/scripts; **tests are canonical** for requirements.
- **D-06:** `--addr` (or env); TLS: default system roots + dev skip-verify; **`context`** for dial/read/handshake cancellation.
- **D-07:** Default **system cert pool**; **explicit** `--insecure-skip-verify` (or equivalent); **no** silent skip.
- **D-08:** **No mTLS** in Phase 8 (`docs/spec/v1/security-assumptions.md`).
- **D-09:** Expose **`Client`** (or equivalent); **sync API** + **`context.Context`** for dial/read/session ops; lifecycle clear vs **`net.Conn`**.
- **D-10:** Errors must distinguish **TLS/IO**, **framing**, **control/PROTOCOL_ERROR (`ErrCode`)**; do not swallow **`pkg/framing.ErrCode`** / **`pkg/protocol`** detail.
- **D-11:** Document fixed **`stream_id`** for demos (suggested **broadcast `1`, unicast `2`**) in package comment + one **`docs/`** location.
- **D-12:** **`testing`** only; table-driven + optional **`testdata/`**; **no** default testify.

### Claude's Discretion
Subcommand literals, package paths (`internal/` split), fake struct names, optional minimal stdin/stdout demo — per plan/execute within D-01..D-12.

### Deferred Ideas (OUT OF SCOPE)
- Two clients + production Relay + CI E2E → **Phase 11 (E2E-01)**.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| **CLNT-01** | TCP+TLS to Relay; **SESSION_CREATE**; receive `session_id` + `invite_code` per spec | `protocol.EncodeSessionCreateReq`, `DecodeSessionCreateAck`; `framing.ParseFrame`/`AppendFrame`; TLS dial + read loop |
| **CLNT-02** | **SESSION_JOIN** by session_id or invite; after **SESSION_JOIN_ACK**, non-zero **`peer_id`** | `protocol.EncodeSessionJoinReq`/`DecodeSessionJoinReq`, `DecodeSessionJoinAck`; session state per `session-state.md` |
| **CLNT-03** | After JOIN_ACK, send/receive **STREAM_DATA** with routing prefix; **repeatable** broadcast + unicast paths; **`stream_id` documented** | `protocol.EncodeStreamData`, `DecodeStreamData`, `RoutingPrefix`, `ValidateRoutingIntent`; fake with two connections or sequential single-connection tests per harness design |
</phase_requirements>

## Executive summary

Phase 8 adds **`pkg/client`**: a small synchronous API over **`crypto/tls.Conn`** (or underlying **`net.Conn`**) that runs the **v1 framing loop** (`pkg/framing`), decodes **payloads** with **`pkg/protocol`**, tracks **joined** state after **`SESSION_JOIN_ACK`**, and enforces **PROT-02** by calling **`JoinGateAllowsBusinessDataPlane`** before accepting or emitting **business** data-plane **`STREAM_*`** traffic (see [join_gate.go](../../pkg/protocol/join_gate.go)). **No `cmd/` tree exists yet** in the repo; Phase 8 introduces a **single `main`** with **client** subcommands for manual smoke only—**automated proof is `go test`** against a **fake peer** that speaks the same bytes on **127.0.0.1** with **TLS** (minimal cert setup), not Docker.

**Primary recommendation:** Implement **`Client`** as: TLS dial → framed read/write loop → explicit **`CreateSession` / `JoinSession`** methods that exchange control frames → post-ACK **`SendStreamData` / `ReadFrame` or `NextMessage`** that only allows **`STREAM_*` sends** when `joined==true`, and surfaces **`framing.ErrCode`** via **`protocol.DecodeProtocolError`** when `msg_type==PROTOCOL_ERROR`. Cover **CLNT-03** with a **two-connection fake** (or sequential harness) that forwards **broadcast** to the other peer and **unicast** by `dst_peer_id`.

## Current codebase anchors (files / functions)

| Area | Location | Role for Phase 8 |
|------|----------|------------------|
| Framing loop | `pkg/framing/decode.go` — `ParseFrame`, `AppendFrame`, `Frame`, `ErrNeedMore`, `ErrFrameTooLarge`, `ErrProtoVersion` | Client read buffer: append TLS reads → loop `ParseFrame` until `ErrNeedMore`; write `AppendFrame` for every send |
| Frame errors | `pkg/framing/errors.go` — `ErrCode`, constants `ErrCode*` | Map `PROTOCOL_ERROR` payloads to these codes via `protocol.DecodeProtocolError` |
| Control messages | `pkg/protocol/control.go` — `EncodeSessionCreateReq`, `DecodeSessionCreateAck`, `EncodeSessionJoinReq`, `DecodeSessionJoinAck`, `EncodeProtocolError`, `DecodeProtocolError` | CREATE/JOIN handshake; parse errors |
| Data plane | `pkg/protocol/streamdata.go` — `EncodeStreamData`, `DecodeStreamData`, `EncodeStreamOpen`/`Close` | CLNT-03 send/receive; optional OPEN/CLOSE for future relay strictness |
| Routing | `pkg/protocol/routing.go` — `RoutingPrefix`, `ParseRoutingPrefix`, `ValidateRoutingIntent`, `RoutingModeBroadcast` / `Unicast` | Build prefixes: broadcast `dst_peer_id==0`, unicast non-zero `dst` |
| Join gate (PROT-02) | `pkg/protocol/join_gate.go` — `JoinGateAllowsBusinessDataPlane` | Before **sending** `STREAM_OPEN`/`STREAM_DATA`/`STREAM_CLOSE` when `!joined`, gate returns not allowed; on **receive**, treat inbound `STREAM_*` before join per fake/relay policy (tests may assert PROTOCOL_ERROR or drop) |
| Msg types | `pkg/protocol/msgtype.go` | Opcode constants for switches |
| Tests pattern | `pkg/protocol/join_gate_test.go` | Table-driven matrices; replicate style in `pkg/client` |
| CLI | **None** — no `cmd/` yet | New: e.g. `cmd/tunnel/main.go` + `internal/cli` or flat `cmd/tunnel/client.go` |

## Protocol flow mapping (client view)

1. **TCP connect** → **`tls.Client`****(TCP conn, tls.Config)** — server name from addr, root CAs from system unless `--insecure-skip-verify` sets `InsecureSkipVerify` **explicitly** (D-07).
2. **Read loop** (per `transport-binding.md`): buffer bytes → **`ParseFrame`** → for each full `Frame`, inspect `f.Payload[0]` (`msg_type`).
3. **SESSION_CREATE (CLNT-01):** send one frame: payload = **`EncodeSessionCreateReq()`** (single-byte `0x01` per implementation). Wait for **`SESSION_CREATE_ACK`**: **`DecodeSessionCreateAck`** → `session_id`, `invite_code`.
4. **SESSION_JOIN (CLNT-02):** **`EncodeSessionJoinReq(joinBy, credential)`** with `join_by=0` + session_id string or `join_by=1` + invite code. Wait for **`SESSION_JOIN_ACK`**: **`DecodeSessionJoinAck`** → set **local `peer_id`**, **`joined=true`**.
5. **STREAM_DATA (CLNT-03):** For each send, set **`RoutingPrefix`** with **`src_peer_id=local peer_id`**, mode **BROADCAST** or **UNICAST**, **`ValidateRoutingIntent`**, then **`EncodeStreamData`**. **Before join:** **`JoinGateAllowsBusinessDataPlane(false, payload)`** must allow only non–data-plane business per gate (STREAM_* blocked for *send*).
6. **Errors:** If payload is **`PROTOCOL_ERROR`**, **`DecodeProtocolError`** yields **`framing.ErrCode`** + reason string.

**Spec nuance (`session-state.md` STATE-01):** client **MUST NOT** send data-plane routing frames before **`SESSION_JOIN_ACK`**. The shared helper **`JoinGateAllowsBusinessDataPlane`** matches that for **`STREAM_*`** opcodes.

## Fake peer design options

| Option | Idea | Pros | Cons |
|--------|------|------|------|
| **A. `net.ListenTCP` + `tls.Server`** in test | Fake relay: **127.0.0.1:0**, self-signed or `crypto/tls` test cert, single goroutine per accepted conn | **Real TLS + real half-close behavior**; matches D-02/D-03 | Must generate/load cert (small helper in test or `testdata/`) |
| **B. Framing-only `net.Conn` (no TLS)** | Pipe or buffer conn, same framing | Simpler | **Violates D-02** (“TLS + … 路径一致”) — use only for unit sub-tests, not primary CLNT proof |
| **C. Two-connection fake registry** | Fake holds `map[peer_id]conn` per session; on **`STREAM_DATA` BROADCAST**, write copy to other peers; on **UNICAST**, lookup `dst_peer_id` | Proves **CLNT-03** broadcast/unicast **without** Phase 9 Relay | Must implement minimal session table + JOIN routing (still not `package relay`) |

**Recommendation:** **A + C** — TLS listener with **minimal session state** (one session_id/invite for test), assign **peer_id** 1 and 2 to two sequential joins (or two parallel dials), forward **`STREAM_DATA`** per routing rules. Package as **`internal/fakepeer`** or `pkg/client/fake_test.go` per plan (**not** `relay`).

**STREAM_OPEN:** v1 suggests **`STREAM_OPEN`** before data; Phase 8 fake may accept **`STREAM_DATA`** alone for simplicity—**document** if fake is looser than future Relay; Phase 10 may require OPEN—track as integration risk (below).

## `pkg/client` API sketch (illustrative)

```go
// Illustrative only — names are discretion (D-09).

type Client struct {
    conn *tls.Conn
    // read buffer for framing
    joined bool
    peerID uint64
}

func Dial(ctx context.Context, addr string, tlsCfg *tls.Config) (*Client, error)

func (c *Client) CreateSession(ctx context.Context) (sessionID, inviteCode string, err error)
func (c *Client) JoinSession(ctx context.Context, bySessionID bool, credential string) (peerID uint64, err error)

func (c *Client) SendStreamData(ctx context.Context, mode uint8, dstPeerID uint64, streamID uint32, flags uint8, app []byte) error
func (c *Client) ReadFrame(ctx context.Context) (framing.Frame, error) // or higher-level NextPayload()

func (c *Client) Close() error
```

**Internals:** single **append-only read buffer**; **`ParseFrame`** in a loop; **`context`** deadlines on `conn.SetDeadline` per op; **send path** checks **`JoinGateAllowsBusinessDataPlane(c.joined, payload)`** before writing **`STREAM_*`**.

## `cmd` subcommands sketch

- **Binary:** e.g. `tunnel` or `tunnel-client` — **one `main` package** (D-05).
- **Flags:** `--addr` (required for client ops), **`--insecure-skip-verify`** (bool, default false), optional **`--timeout`** mapping to **context** deadline.
- **Subcommands (examples):** `client create` → print `session_id`, `invite_code`; `client join --session <id>` or `--invite <code>` → print `peer_id`; optional `client send`/`recv` for smoke — **not** the canonical CLNT verifier.

## TLS and `context`

- **Default:** `tls.Config{ServerName: host}` + **nil `RootCAs`** (uses system pool) — D-07.
- **Dev:** user passes **`InsecureSkipVerify: true`** only via **explicit CLI flag** — never default.
- **SNI / host:** parse `host:port` from `--addr` for `ServerName` when connecting by IP literal may need careful `ServerName` handling (document: use hostname in addr for cert validation when possible).
- **`context`:** `DialContext` if using `net.Dialer`; for **`Read`/`Write`**, set deadlines from `ctx.Deadline()` or use `conn.SetDeadline` before each blocking op so **`ctx.Done()`** cancels promptly (or run reads in goroutine with `select`—keep simple for v1).

## Error model

| Layer | Source | Client behavior |
|-------|--------|------------------|
| TLS / TCP | `tls.AlertError`, `net.OpError`, `io.EOF` | Wrap or return as **`ClientError`** kind `Transport` (implementation naming open) |
| Framing | `ErrNeedMore` (internal), `ErrFrameTooLarge`, `ErrProtoVersion` | Expose as **`FramingError`** / sentinel |
| Protocol | `DecodeProtocolError` → `framing.ErrCode` + reason | **`ProtocolError`** carrying **`framing.ErrCode`** + reason string — **D-10** |
| Join gate | Local `JoinGateAllowsBusinessDataPlane` == false for send | **`errors.New("protocol: ...")`** or typed **`ErrNotJoined`** before wire write |

Use **`errors.Is` / `errors.As`** or small **typed errors** so callers never lose **`ErrCode`**.

## Testing strategy

- **Unit:** framing buffer edge cases (empty, multiple frames, oversized) — table-driven, optional **`testdata/`** binary blobs (**D-12**).
- **Integration (primary):** `go test` starts **TLS fake** on **`127.0.0.1:0`**:
  - **CLNT-01:** one client **`CreateSession`** → assert UUID + invite format (match **`protocol`** validators).
  - **CLNT-02:** **`JoinSession`** → non-zero **`peer_id`**.
  - **CLNT-03:** two **`Client`**s same fake session → **broadcast** (`RoutingModeBroadcast`, `stream_id=1` per D-11) observed on peer B; **unicast** (`RoutingModeUnicast`, `dst_peer_id=B`, `stream_id=2`) only on B.
- **Join gate:** optional test that **pre-ACK `SendStreamData`** fails locally without emitting forbidden bytes (or fake receives **PROTOCOL_ERROR** if client mistakenly sends—prefer **client-side refusal** aligned with PROT-02).

## Risks

| Risk | Mitigation |
|------|------------|
| **Fake diverges from future Relay** | Keep fake strictly **frame-compatible** with **`pkg/protocol`**; minimize magic; document differences (e.g. OPEN optional). |
| **`stream_id` 0 invalid** (`streams-lifecycle.md`) | Use documented **1** and **2** (D-11); never **0** in tests. |
| **Cert / TLS flakiness in CI** | Generate cert once in test setup (`tls.Certificate` from `tls.X509KeyPair` or embedded PEM in `testdata/`). |
| **Blocking reads vs `ctx`** | Always set **read deadline** from context; document that **`Close()`** unblocks reads. |
| **Package name collision** | Never **`package relay`** for fake (**D-04**). |

## Validation Architecture (Nyquist)

`.planning/config.json` has **`workflow.nyquist_validation`: true** — Phase 8 must define **automated** checks for **CLNT-01..03**.

### Test framework

| Property | Value |
|----------|-------|
| Framework | Go **`testing`** (stdlib) |
| Config file | none — `go test ./...` |
| Quick run | `go test ./pkg/client/... -count=1` |
| Full suite | `go test ./... -count=1` |

### Phase requirements → test map

| Req ID | Behavior | Test type | Automated command | Notes |
|--------|----------|-----------|-------------------|--------|
| **CLNT-01** | TCP+TLS dial; **SESSION_CREATE**; parse **`session_id`**, **`invite_code`** | integration (fake TLS server) | `go test ./pkg/client -run TestClient_CreateSession -count=1` | Assert decoded strings match fake + **`protocol.DecodeSessionCreateAck`** |
| **CLNT-02** | **SESSION_JOIN**; **SESSION_JOIN_ACK** → **`peer_id` != 0** | integration | `go test ./pkg/client -run TestClient_JoinSession -count=1` | Optionally table-driven: join by session_id vs invite_code |
| **CLNT-03** | Post-ACK **STREAM_DATA** send/recv; **broadcast** and **unicast** paths | integration (two clients or harness) | `go test ./pkg/client -run TestClient_StreamData -count=1` | Fixed **`stream_id`** 1/2 per D-11; assert **`RoutingPrefix`** on receive |

### Sampling rate

- **Per task / commit:** `go test ./pkg/client/... -short` (if `-short` used) or full **`pkg/client`** tests.
- **Phase gate:** **`go test ./...`** green before **`/gsd-verify-work`**.

### Wave 0 gaps

- [ ] **`pkg/client`** package + `_test.go` with TLS fake — **covers CLNT-01..03**
- [ ] **`cmd/...`** may have **no** automated requirement tests — acceptable if **D-05** (CLI auxiliary); document manual smoke only
- [ ] **Cert helper** shared by tests — small **`testdata/`** PEM or dynamic cert generation

*(If implementation places tests under `internal/fakepeer`, adjust `-run` paths but keep **`go test ./...` as the gate**.)*

## Sources

### Primary (HIGH confidence)
- `pkg/framing/decode.go`, `pkg/framing/errors.go`
- `pkg/protocol/*.go` (control, streamdata, routing, join_gate)
- `docs/spec/v1/session-create-join.md`, `routing-modes.md`, `streams-lifecycle.md`, `transport-binding.md`, `errors.md`, `session-state.md`
- `08-CONTEXT.md` (D-01..D-12)

### Secondary (MEDIUM confidence)
- Go `crypto/tls` package docs (local): standard library behavior for **`InsecureSkipVerify`**, **`ServerName`**

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | build/test | ✓ | go1.22+ (repo); dev machine **1.26.1** observed | — |
| `go test` | Nyquist / CI | ✓ | — | — |
| Docker | — | ✗ | — | **Not used** (D-03) |

**Missing dependencies with no fallback:** none for planned approach.

## Metadata

**Research date:** 2026-04-04  
**Valid until:** ~30 days (stable stack); revisit if **`pkg/protocol` public API changes post–Phase 7.

## RESEARCH COMPLETE
