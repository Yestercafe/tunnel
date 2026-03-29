package framing

import "testing"

// Golden values must match docs/spec/v1/errors.md (ERR-01).
// Phase 6：与 errors.md 对照无新增码。
func TestErrCode_matchesSpecTable(t *testing.T) {
	cases := []struct {
		name string
		code ErrCode
		want uint16
	}{
		{"ERR_FRAME_TOO_LARGE", ErrCodeFrameTooLarge, 0x0001},
		{"ERR_PROTO_VERSION", ErrCodeProtoVersion, 0x0002},
		{"ERR_JOIN_DENIED", ErrCodeJoinDenied, 0x0003},
		{"ERR_SESSION_NOT_FOUND", ErrCodeSessionNotFound, 0x0004},
		{"ERR_ROUTING_INVALID", ErrCodeRoutingInvalid, 0x0005},
		{"ERR_ENVELOPE_INVALID", ErrCodeEnvelopeInvalid, 0x0006},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if uint16(tc.code) != tc.want {
				t.Fatalf("ErrCode %s: got 0x%04x, want 0x%04x", tc.name, tc.code, tc.want)
			}
		})
	}
}
