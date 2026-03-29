package appenvelope

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSplitApplicationData(t *testing.T) {
	cases := []struct {
		name       string
		flags      uint8
		app        []byte
		wantEnv    []byte
		wantBody   []byte
		wantErr    error
		wantErrSub string
	}{
		{
			name:     "no flag body all opaque",
			flags:    0x00,
			app:      []byte{1, 2, 3},
			wantEnv:  nil,
			wantBody: []byte{1, 2, 3},
		},
		{
			name:       "has envelope but too short",
			flags:      FlagHasAppEnvelope,
			app:        []byte{0},
			wantErr:    ErrEnvelopeTooShort,
			wantErrSub: "",
		},
		{
			name:     "has envelope len zero",
			flags:    0x02,
			app:      []byte{0x00, 0x00},
			wantEnv:  []byte{},
			wantBody: []byte{},
		},
		{
			name:       "has envelope truncated",
			flags:      0x02,
			app:        []byte{0x00, 0x05, 1, 2},
			wantErr:    ErrEnvelopeTruncated,
			wantErrSub: "",
		},
		{
			name:     "has envelope len 2 body rest",
			flags:    0x02,
			app:      []byte{0x00, 0x02, 0xAB, 0xCD, 9, 9},
			wantEnv:  []byte{0xAB, 0xCD},
			wantBody: []byte{9, 9},
		},
		{
			name:     "FIN and HAS same split as HAS only",
			flags:    0x03,
			app:      []byte{0x00, 0x01, 0xEE, 0xFF},
			wantEnv:  []byte{0xEE},
			wantBody: []byte{0xFF},
		},
		{
			name:       "envelope_len over max",
			flags:      FlagHasAppEnvelope,
			app:        append([]byte{0x10, 0x01}, bytes.Repeat([]byte{'x'}, 4097)...),
			wantErrSub: "exceeds max",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env, body, err := SplitApplicationData(tc.flags, tc.app)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if tc.wantErrSub != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("err = %v, want substring %q", err, tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(env, tc.wantEnv) {
				t.Fatalf("envelope = %v, want %v", env, tc.wantEnv)
			}
			if !bytes.Equal(body, tc.wantBody) {
				t.Fatalf("body = %v, want %v", body, tc.wantBody)
			}
		})
	}
}
