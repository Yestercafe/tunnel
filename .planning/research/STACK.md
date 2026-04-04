# 技术栈研究：最小 Relay + Client（Go，TCP+TLS，v1 成帧）

**领域：** 在 TLS 字节流上承载 v1 逻辑帧的中继与客户端  
**调研日期：** 2026-04-04  
**置信度：** **HIGH**（标准库与官方发布说明）；第三方依赖为 **N/A**（本里程碑推荐零新增模块依赖）

## Recommended Stack

### Core Technologies

| 组件 | 版本 | 用途 | 推荐理由 |
|------|------|------|----------|
| **Go toolchain** | `go 1.22`（`go.mod` 已声明） | 语言与运行时 | 与仓库约束及既有 `pkg/framing` / CI 一致；Relay/Client 所需网络与 TLS 能力均在标准库内完备。 |
| **`net`** | 随上述 Go 发行版 | `Listen` / `Dial`、每连接 `net.Conn`、地址解析 | 最小 Relay：`TCPListener` 接受连接后再进入 TLS；Client：`Dial`/`DialContext` 建链。与 v1「TCP 字节流」一致。 |
| **`crypto/tls`** | 随上述 Go 发行版 | `tls.Server` / `tls.Client`、`tls.Config`、ALPN（若需要可留空） | v1 假设 **TLS 在边缘/服务器终止**；最小实现即在进程内终止 TLS，在 `*tls.Conn` 上读写字节流，再交给成帧层。官方包文档见 [crypto/tls](https://pkg.go.dev/crypto/tls)。 |
| **`context`** | 随上述 Go 发行版 | 监听关闭、拨号超时、连接生命周期取消 | 与 `net.DialContext`、`Listen` 关闭协作，避免 goroutine 泄漏。 |

### Supporting Libraries

| 库 | 版本 | 用途 | 何时使用 |
|----|------|------|----------|
| **`bufio`**（标准库） | 随 Go | 可选：对 `tls.Conn` 做读缓冲 | **可选**。`pkg/framing.ParseFrame` 面向 `[]byte`；若自管缓冲区则未必需要 `bufio`；若希望统一 `Read` 块大小、简化循环，可用 `bufio.Reader` 仍从底层取字节填入自定义 buffer。**不新增模块依赖。** |
| **`sync`**（标准库） | 随 Go | `Mutex` / `RWMutex`、`WaitGroup` | session ↔ peers 映射、每连接 handler 与全局注册表之间的并发安全。 |
| **`crypto/x509` + `encoding/pem` + `crypto/rand` + `crypto/rsa` 或 `ecdsa`** | 随 Go | 测试或本地演示用的自签证书 | **仅测试/演示**：与 `tls.Config.Certificates` / `tls.X509KeyPair` 配合；亦可改为仓库 `testdata/` 固定 PEM（与现有 golden/测试风格一致），二者择一即可。 |
| **（无）第三方成帧 / TLS 封装** | — | — | **不要引入**：应用层帧格式已由 **`pkg/framing`**（`ParseFrame` / `AppendFrame`）覆盖。 |

### Development Tools

| 工具 | 用途 | 说明 |
|------|------|------|
| **`go test ./...`** | 与现有一致性测试、未来 E2E 同进程测 | 已在 CI 中使用；Relay/Client 测例可用 `net.Listener` + 内存生成或 `testdata` 证书搭建本地 TLS，无需新测试框架。 |
| **官方发布页** | 核实 **1.22.x 最新补丁** | 安全修复随补丁发布；请在 [Go Release History](https://go.dev/doc/devel/release.html) 查询 **当前 1.22 系列最新版本** 作为 CI/本机建议下限（**勿在文档中写死易过时的补丁号**，以该页为准）。 |

## Installation

本里程碑 **无需** `go get` 新增依赖；保持 `go.mod` 仅含：

```bash
# 已满足：模块声明 go 1.22，无 require 块亦可
go version    # 建议使用 release.html 上 go1.22 系列当前最新补丁
go test ./...
```

若日后仅为测试人体工程学引入断言库（如 `testify`），属**可选**，且非 v1.1 最小闭环所必需。

## Alternatives Considered

| 推荐 | 替代方案 | 何时考虑替代 |
|------|----------|--------------|
| 标准库 `crypto/tls` | `github.com/quic-go/quic-go`（QUIC）等 | **不适用 v1**：规范与里程碑锁定 **TCP+TLS 字节流**。 |
| 标准库 `net` + 手写读缓冲 + `framing.ParseFrame` | 第三方 length-prefix 库 | **不推荐**：与既有 `pkg/framing` 重复，且协议已固定。 |
| `sync.RWMutex` 保护 session 表 | `sync.Map` | 读多写少且键为 `session_id` 时可评估 `sync.Map`；最小实现先用 `map+Mutex` 更清晰。 |

## What NOT to Use

| 避免 | 原因 | 改用 |
|------|------|------|
| **WebSocket 栈**（如 `gorilla/websocket`） | v1 明确不以 WebSocket 为承载；规范与测试均基于 TLS 字节流成帧 | `TCP` + `crypto/tls` + `pkg/framing` |
| **gRPC / Connect 等 RPC 框架** | 引入另一套 framing 与语义，与 v1 控制面/数据面帧布局无关 | 直接在 `Frame.Payload` 上实现 SESSION_CREATE / JOIN / 数据路由 |
| **额外 TLS 封装库**（如 `utls`） | 指纹伪装等非 v1.1 目标 | `crypto/tls` |
| **可观测性全家桶**（OTel、Prometheus 客户端等） | v1.1 **最小**可运行；非表功能 | 里程碑后期或运维化阶段再加 |

## Stack Patterns by Variant

**若实现最小 Relay（监听 + TLS + 每连接成帧）：**

- 使用 `net.Listen("tcp", addr)`，对该 `net.Listener` 使用 `tls.NewListener(inner, serverTLSConfig)` **或** 在 `Accept` 后对每连接执行 `tls.Server(conn, config)`（两种均为常见模式，择一保持代码风格统一）。
- 每连接一个循环：从 `*tls.Conn` **读入字节**追加到缓冲区，对缓冲区调用 `framing.ParseFrame`；`ErrNeedMore` 则继续 `Read`；解析成功后**前移缓冲区**或使用 slice 窗口，避免重复解析。
- 写出路径：控制面/数据面构造 `framing.Frame` 后 **`AppendFrame` → `Write`/`Write` 全量发出**（必要时处理短写循环）。

**若实现最小 Client（拨号 TLS + session 创建/加入 + 广播/单播）：**

- 使用 `tls.Dial` / `tls.DialWithDialer` + `context` 控制超时；`tls.Config` 中设置 `ServerName`（SNI）与可信根（生产）；本地测试可用 `InsecureSkipVerify`（**仅限测试**）或私有 CA。
- 与 Relay **共用同一套**：`pkg/framing` 编码/解码与 `pkg/appenvelope`（若 payload 需 JSON 信封）的拼装规则；不在此里程碑引入新序列化框架。

## Version Compatibility

| 组件 | 兼容关系 | 说明 |
|------|----------|------|
| `go 1.22`（`go.mod`） | `pkg/framing`、`pkg/appenvelope` 现有代码 | 无额外版本矩阵；升级 Go 主版本时按 [Go 1 compatibility](https://go.dev/doc/go1compat) 评估。 |
| `crypto/tls` | TLS 1.2+ 默认协商 | 最小实现可显式 `MinVersion: tls.VersionTLS12`；需与运维/规范中的「TLS 在边缘」假设一致（**协议内无 E2E**，见 `security-assumptions.md`）。 |
| **补丁版本** | 任意 **go1.22.x** 与 `go.mod` 兼容 | **具体补丁号**以 [release.html](https://go.dev/doc/devel/release.html) 最新 **go1.22.\*** 为准（调研时 1.22 系列已有多个安全修复版本，**请在 CI 镜像中定期对齐**）。 |

## 与现有 `pkg/framing` 的集成要点

- **读路径**：`ParseFrame(buf []byte)` 为**同步块解析**；TCP+TLS 为流式输入，必须在连接状态内维护**累积 buffer**（或 ring buffer），与 `ErrNeedMore` 配合，直至完整帧再交给会话/路由逻辑。
- **写路径**：`AppendFrame` 已含固定头 + payload 长度；与 v1 规范一致，**无需**再包一层长度前缀。
- **错误语义**：`pkg/framing` 已提供与规范对齐的错误类型；Relay 对协议错误应能映射到规范中的 `PROTOCOL_ERROR` 等（实现阶段按 `docs/spec/v1/errors.md` 与现有 `ErrCode`）。

## Sources

- [Go Release History](https://go.dev/doc/devel/release.html) — go1.22.x 补丁线与安全修复（**HIGH**）
- [crypto/tls - pkg.go.dev](https://pkg.go.dev/crypto/tls) — TLS 服务端/客户端 API（**HIGH**）
- 仓库 `go.mod`、`pkg/framing/decode.go` — 成帧 API 与模块约束（**HIGH**）
- `.planning/PROJECT.md` — v1.1 范围与 Out of Scope（**HIGH**）

---
*Stack research for: Tunnel v1.1 最小 Relay + Client*  
*Researched: 2026-04-04*
