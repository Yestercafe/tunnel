// Package appenvelope splits STREAM_DATA application_data per APP-01 (optional UTF-8 JSON envelope).
// See docs/spec/v1/app-envelope.md — byte boundaries only; JSON semantics are handled by callers.
package appenvelope

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// FlagFIN is STREAM_DATA flags bit 0 (FIN).
	FlagFIN = 1 << 0
	// FlagHasAppEnvelope is STREAM_DATA flags bit 1 (HAS_APP_ENVELOPE).
	FlagHasAppEnvelope = 1 << 1
)

// Well-known decode errors for application_data boundaries.
var (
	ErrEnvelopeTooShort = errors.New("application_data too short for envelope_len prefix")
	ErrEnvelopeTruncated = errors.New("application_data truncated: envelope_len exceeds remainder")
)

// MaxEnvelopeLen is the v1 maximum envelope_len (APP-01).
const MaxEnvelopeLen = 4096

// SplitApplicationData splits application_data from flags and payload bytes.
// When HAS_APP_ENVELOPE is clear, envelope is nil and body is the full slice.
// When set, bytes [0:2] are uint16 BE envelope_len, then envelope, then body.
func SplitApplicationData(flags uint8, applicationData []byte) (envelope []byte, body []byte, err error) {
	if flags&FlagHasAppEnvelope == 0 {
		return nil, applicationData, nil
	}
	if len(applicationData) < 2 {
		return nil, nil, ErrEnvelopeTooShort
	}
	n := int(binary.BigEndian.Uint16(applicationData[0:2]))
	if n > MaxEnvelopeLen {
		return nil, nil, fmt.Errorf("envelope_len %d exceeds max %d", n, MaxEnvelopeLen)
	}
	if 2+n > len(applicationData) {
		return nil, nil, ErrEnvelopeTruncated
	}
	envelope = applicationData[2 : 2+n]
	body = applicationData[2+n:]
	return envelope, body, nil
}
