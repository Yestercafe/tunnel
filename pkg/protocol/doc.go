// Package protocol implements the v1 payload semantic layer for control and data-plane messages.
//
// It does not parse the 10-byte framing header from pkg/framing; callers pass frame payloads only.
//
// Requirements covered here include PROT-01 (control + STREAM_* field views) and PROT-02
// (join_gate — see JoinGateAllowsBusinessDataPlane). PROT-02 aligns with docs/spec/v1/session-state.md
// STATE-01: before SESSION_JOIN_ACK with a valid peer_id, peers MUST NOT treat data-plane STREAM_*
// frames as legitimate business traffic; PROTOCOL_ERROR (0x05) is control-plane and is not
// conflated with STREAM_* opcodes.
package protocol
