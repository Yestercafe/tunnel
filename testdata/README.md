<!-- REQ TEST-01 -->

# `testdata/` 约定（Phase 6 / 一致性测试）

本目录存放 **golden 向量** 与 **负例** 输入，供 `go test` 从文件加载，便于审阅与增量扩展。成帧与载荷分层见 `docs/spec/v1/frame-layout.md`（FRAME-01）；应用信封边界见 `docs/spec/v1/app-envelope.md`（APP-01）。

## 子目录

| 目录 | 用途 |
|------|------|
| `framing/` | **成帧层** golden：完整逻辑帧或仅帧头前缀（负例/半包） |
| `appenvelope/` | **应用信封切分**（`SplitApplicationData`）向量 |

## 文件扩展名

| 扩展名 | 含义 |
|--------|------|
| `.hex` | **十六进制文本**：允许空格与换行；测试侧用 `strings.Fields` 拼接后 `hex.DecodeString` 得到原始字节 |
| `.bin` | **原始二进制**（与 `.hex` 二选一，按场景选用） |
| `.hextxt` | **行分隔的文本驱动用例**（仅 `appenvelope` 分割测试使用）：与 `.hex`（纯十六进制字节块）区分，格式见该目录内文件头注释 |

## 命名建议

`{scenario}_{want}.hex`，例如 `empty_payload_ok.hex`、`frame_too_large_prefix.err.hex`。

## 规范引用

- 帧布局：`docs/spec/v1/frame-layout.md`
- 应用信封：`docs/spec/v1/app-envelope.md`
