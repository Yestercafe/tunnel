package framing

// ErrCode is a v1 application-layer error code (uint16, big-endian in PROTOCOL_ERROR payloads).
// Symbol names and uint16 values match docs/spec/v1/errors.md (ERR-01).
type ErrCode uint16

const (
	// ErrCodeFrameTooLarge corresponds to ERR_FRAME_TOO_LARGE in docs/spec/v1/errors.md.
	ErrCodeFrameTooLarge ErrCode = 0x0001
	// ErrCodeProtoVersion corresponds to ERR_PROTO_VERSION in docs/spec/v1/errors.md.
	ErrCodeProtoVersion ErrCode = 0x0002
	// ErrCodeJoinDenied corresponds to ERR_JOIN_DENIED in docs/spec/v1/errors.md.
	ErrCodeJoinDenied ErrCode = 0x0003
	// ErrCodeSessionNotFound corresponds to ERR_SESSION_NOT_FOUND in docs/spec/v1/errors.md.
	ErrCodeSessionNotFound ErrCode = 0x0004
	// ErrCodeRoutingInvalid corresponds to ERR_ROUTING_INVALID in docs/spec/v1/errors.md.
	ErrCodeRoutingInvalid ErrCode = 0x0005
	// ErrCodeEnvelopeInvalid corresponds to ERR_ENVELOPE_INVALID in docs/spec/v1/errors.md.
	ErrCodeEnvelopeInvalid ErrCode = 0x0006
)
