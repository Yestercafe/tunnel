package appenvelope

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
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

// TestSplitApplicationData_FileDriven loads testdata/appenvelope/split_cases.hextxt (docs/spec/v1/app-envelope.md).
func TestSplitApplicationData_FileDriven(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "appenvelope", "split_cases.hextxt")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 5 {
			t.Fatalf("bad row (want 5 TAB-separated columns): %q", line)
		}
		flagsB, err := hex.DecodeString(strings.Join(strings.Fields(parts[0]), ""))
		if err != nil || len(flagsB) != 1 {
			t.Fatalf("flags %q: %v", parts[0], err)
		}
		app, err := hex.DecodeString(strings.Join(strings.Fields(parts[1]), ""))
		if err != nil {
			t.Fatalf("app %q: %v", parts[1], err)
		}
		kind := strings.TrimSpace(parts[2])
		wantEnv := decodeOptionalHex(t, parts[3])
		wantBody := decodeOptionalHex(t, parts[4])

		env, body, gotErr := SplitApplicationData(flagsB[0], app)
		switch kind {
		case "ok":
			if gotErr != nil {
				t.Fatalf("row %q: err = %v", line, gotErr)
			}
			if !bytesEqualNilEmpty(env, wantEnv) {
				t.Fatalf("row %q: env %v want %v", line, env, wantEnv)
			}
			if !bytes.Equal(body, wantBody) {
				t.Fatalf("row %q: body %v want %v", line, body, wantBody)
			}
		case "ErrEnvelopeTooShort":
			if !errors.Is(gotErr, ErrEnvelopeTooShort) {
				t.Fatalf("row %q: err = %v want ErrEnvelopeTooShort", line, gotErr)
			}
		case "ErrEnvelopeTruncated":
			if !errors.Is(gotErr, ErrEnvelopeTruncated) {
				t.Fatalf("row %q: err = %v want ErrEnvelopeTruncated", line, gotErr)
			}
		default:
			t.Fatalf("unknown want_kind %q in %q", kind, line)
		}
	}
}

func decodeOptionalHex(t *testing.T, s string) []byte {
	t.Helper()
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return nil
	}
	b, err := hex.DecodeString(strings.Join(strings.Fields(s), ""))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func bytesEqualNilEmpty(a, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return bytes.Equal(a, b)
}
