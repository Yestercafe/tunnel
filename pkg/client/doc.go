// Package client implements a v1 Tunnel client over TCP+TLS with framing + protocol payloads.
//
// Demo stream_id policy (see docs/client-stream-ids.md):
//   - stream_id 1 — broadcast demo payloads
//   - stream_id 2 — unicast demo payloads
//
// stream_id 0 is invalid per docs/spec/v1/streams-lifecycle.md (STREAM-02).
package client
