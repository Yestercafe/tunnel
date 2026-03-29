# Phase 6: 一致性测试套件 — Research

**Date:** 2026-03-29  
**Question:** 如何以 Go + `testdata/` + CI 满足 TEST-01 与 ROADMAP Phase 6，并与现有 `pkg/framing`、`pkg/appenvelope` 对齐？

---

## Summary

- **载体：** 仓库根 `testdata/framing/`、`testdata/appenvelope/`（及可选 `testdata/integration/`）存放 `.hex`（去空格换行后 `hex.DecodeString`）或 `.bin`；与 `06-CONTEXT.md` D-01～D-02 一致。  
- **第一层：** 扩展 `pkg/framing`：将 `decode_test.go` 中已有十六进制迁入 `testdata/` 表驱动；补充 `ErrNeedMore`（短缓冲）、`ErrProtoVersion`（版本≠0x0001）、与 `errors_test.go` 中 `ErrCode` 表一致性。  
- **第二层：** `pkg/appenvelope`：对 `SplitApplicationData` 用 `testdata` 行（flags + hex app）补充与 `split_test.go` 等价覆盖，减少重复表体积。  
- **第三层（少而精）：** 使用 `routing-modes.md` 广播示例整帧（2 行 hex，去空格拼接）作为 **golden**：仅断言 `ParseFrame` 成功、`payload_len`、payload 前若干字节（路由前缀 + `stream_id`）与文档一致；**不**要求本阶段实现完整 Relay 状态机。  
- **CI：** `.github/workflows/go.yml`，`go test ./... -count=1`，Go **1.22** 与 `go.mod` 一致。  
- **错误路径：** 优先与已有 `ErrCode`、`ParseFrame` 错误映射；若需 `PROTOCOL_ERROR` 载荷编码，与 `docs/spec/v1/errors.md` 同提交引入最小编码函数（CONTEXT D-09）。

---

## Existing Code Anchors

| 区域 | 文件 | 用途 |
|------|------|------|
| 成帧 | `pkg/framing/decode.go` | `ParseFrame` / `AppendFrame` |
| 成帧测试 | `pkg/framing/decode_test.go` | 往返、过大 payload |
| 错误码 | `pkg/framing/errors.go`, `errors_test.go` | `ErrCode` 与规范表 |
| 信封 | `pkg/appenvelope/*.go`, `split_test.go` | `SplitApplicationData` |

---

## Golden 向量来源（规范文档）

| 文档 | 用途 |
|------|------|
| `docs/spec/v1/frame-layout.md` | 帧头 10 字节、示例 hex |
| `docs/spec/v1/routing-modes.md` | 广播整帧示例（`payload_len=22` 等） |
| `docs/spec/v1/streams-lifecycle.md` | `STREAM_DATA` 布局、`flags` 位 |
| `docs/spec/v1/app-envelope.md` | `HAS_APP_ENVELOPE`、`envelope_len` |
| `docs/spec/v1/errors.md` | `ErrCode` 数值 |

**规范示例帧（广播）：** `routing-modes.md` 中两行十六进制拼接后为完整帧字节序列；测试中可固定期望 `len(Payload)==22` 且前缀字节与文档表格一致。

---

## Test Helpers（建议）

- **`testing` 仅**：`filepath.Join` + `os.ReadFile` 读 `testdata`；或 `//go:embed` 跳过（D-03）。  
- **Hex 文件格式：** 允许空格与换行（`strings.Fields` 拼接后 decode）。  
- **负例：** 单独文件 `*_want_err.txt` 或与表列 `want: ErrFrameTooLarge` 同文件命名约定（计划在 06-01 锁定一种）。

---

## Risks & Mitigations

| 风险 | 缓解 |
|------|------|
| 向量与实现漂移 | 每向量在注释中标注 `docs/...` 锚点；CI 必跑 `go test ./...` |
| Phase 6 范围膨胀 | 第三层仅 1～2 个端到端 golden；不测 TLS/TCP |
| APP-01 在 REQUIREMENTS 仍 Pending | 信封测试对齐 **已写** `app-envelope.md` 与 `pkg/appenvelope`，不阻塞 TEST-01 |

---

## Validation Architecture

**维度 1（单元 / 表驱动）：** `pkg/framing`、`pkg/appenvelope` 包内 `_test.go` + `testdata/` 文件驱动；`go test ./... -count=1` 全绿。

**维度 2（golden 文件）：** `testdata/**/*.hex` 或 `.bin` 可 diff；新增场景复制规范示例 hex。

**维度 3（回归门禁）：** GitHub Actions 在 PR/push 上运行相同命令；无 `-race` 强制（CONTEXT D-08）。

**维度 4（可追溯）：** 每个 golden 测试在 `read_first` 或注释中引用 REQ **TEST-01** 与对应规范路径。

**维度 5（负例）：** 至少覆盖 `ErrNeedMore`、`ErrFrameTooLarge`、`ErrProtoVersion`、信封 `ErrEnvelopeTooShort` / `ErrEnvelopeTruncated` 各一类（可表驱动）。

**维度 6（不测范围）：** 不显式断言 TLS alert、TCP 行为（CONTEXT D-10）。

**维度 7（执行采样）：** 每任务提交后运行 `go test ./...`；每 wave 末全量相同。

**维度 8（Nyquist）：** 本阶段「反馈信号」为测试退出码 + 可选 `go test -v` 对新增用例名的 grep；无独立 UI。

---

## RESEARCH COMPLETE

本研究足以支撑 **06-01**（布局与约定）与 **06-02**（解析/信封/少量端到端 golden + CI）的 PLAN 编写。
