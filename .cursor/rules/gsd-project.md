<!-- gsd-project-start source:PROJECT.md -->
## Project

**Tunnel（传输隧道）**

面向公网部署的**中继隧道**：多个 **client** 通过同一服务在同一 **session** 内交换数据。创建者**开房**并获得 `session_id`（或邀请码），其它 **peer** 凭此加入。协议优先：先交付**帧格式、状态机、错误码**与**一致性测试**；实现语言为 **Go**，但第一步不强制完成端到端产品实现。

**面向谁：**需要在公网与内网之间、**进程/服务**之间安全转发数据的开发者与小团队（**少量固定 peer**）。v1 **不以 WebSocket 为承载**；浏览器若需接入见 REQUIREMENTS v2。

**Core Value:** 在 **TLS 由边缘/服务器终止** 的前提下，用**同一套协议**同时支撑：**广播**、**私信（单播）**、**双向流**、**小消息与大块流**；并通过**可选应用信封**让 Web 前端、Copilot 管道等上层复用，而无需各自定义一套私有帧格式。

### Constraints

- **语言**：实现默认 **Go**（规范与测试可先于完整实现落地）。
- **交付顺序**：先**规范 + 一致性测试**，再展开实现与运维化。
- **传输**：v1 **不采用 WebSocket**；主路径为 **TLS（如基于 TCP）上的自定义成帧**。浏览器侧非 v1 目标（见 REQUIREMENTS v2）。
- **安全边界**：机密性与完整性主要依赖 **TLS（边缘/服务器）**；协议负责会话、成员与消息语义。
- **顺序模型**：**流内有序、流间可乱序**（在规范中用流 ID 或逻辑通道精确定义）。
<!-- gsd-project-end -->

<!-- gsd-stack-start source:research/STACK.md -->
## Technology Stack

## Recommended Stack
### Core Technologies
| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.22+（建议跟踪当前稳定版） | 协议参考实现、一致性测试 runner | 静态编译、并发原语成熟、适合网络与二进制协议 |
| `crypto/tls` + `net` | 标准库 | 公网 **TLS over TCP** 双向流 | 与 v1「连续字节流成帧」一致 |
| `github.com/stretchr/testify` | 1.9+ | 断言与测试套件组织 | Go 生态事实标准，适合一致性测试 |
### Supporting Libraries
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `bufio` / `encoding/binary` | 标准库 | 定长与变长字段解析 | 帧解析与 golden test |
| `github.com/google/go-cmp` | 0.6+ | 深度比较结构体 | 复杂消息与状态机测试 |
### Development Tools
| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` | 运行一致性测试 | 可将向量放在 `testdata/` |
| `golangci-lint` | 静态检查 | 避免低级错误进入协议实现 |
## Installation
## Alternatives Considered
| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| TCP+TLS 上自定义成帧 | WebSocket 承载 | 需要浏览器原生 client 时，可在 v2 加适配层 |
| 二进制帧自定义 | 仅用 JSON over TLS | 大载荷与流式场景下带宽与解析成本更高 |
## What NOT to Use
| Avoid | Why | Use Instead |
|-------|-----|-------------|
| 在规范未定前锁死某一「多路复用库」 | 与自研帧/流 ID 模型可能冲突 | 规范中定义流抽象，实现再选库或自研 mux |
| 将 TLS 细节塞进 v1 核心帧 | 与「TLS 在边缘」决策冲突 | 在部署文档与能力位中扩展 |
## Stack Patterns by Variant
- v1 以 **TLS 字节流**为主路径；逻辑帧编码与传输层解耦。
- 若未来需要浏览器：在 **WebSocket 或 WebTransport** 上增加适配章节，**逻辑帧**保持不变。
## Version Compatibility
| 组件 | 兼容说明 |
|------|----------|
| Go 版本 | 建议在 `go.mod` 标明 `go 1.22` 或更新，便于使用新标准库特性 |
| 可选 WS 适配（v2+） | 若增加，需在规范中登记子协议或分帧与逻辑帧的映射 |
## References
- Go 官方文档：Testing、Fuzzing（可用于生成随机帧向量）
<!-- gsd-stack-end -->

<!-- gsd-conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- gsd-conventions-end -->

<!-- gsd-architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- gsd-architecture-end -->

<!-- gsd-workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- gsd-workflow-end -->



<!-- gsd-profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- gsd-profile-end -->
