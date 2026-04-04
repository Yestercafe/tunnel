// Package relay implements the Tunnel v1 Relay server: TCP listen, TLS termination,
// per-connection framing (pkg/framing), and session control (SESSION_CREATE / SESSION_JOIN).
// It is the production relay (not internal/fakepeer). Data-plane STREAM_DATA routing is Phase 10 (RLY-03).
package relay
