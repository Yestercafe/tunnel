# Client demo `stream_id` policy

This repository uses fixed demo values for automated tests and CLI smoke runs:

| Use | `stream_id` |
|-----|----------------|
| Broadcast demo payloads | **1** |
| Unicast demo payloads | **2** |

`stream_id` **0** is invalid per [streams-lifecycle.md](spec/v1/streams-lifecycle.md) (**STREAM-02**).

See also `pkg/client` package documentation and Phase 8 context (`.planning/phases/08-client-pkg-client-cmd/08-CONTEXT.md`).
