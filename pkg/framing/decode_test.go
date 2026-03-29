package framing

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustDecodeHexFile loads a .hex file: strips # comments and whitespace, then hex-decodes.
// See testdata/README.md (REQ TEST-01) and docs/spec/v1/frame-layout.md (FRAME-01).
func mustDecodeHexFile(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "framing", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sb.WriteString(strings.Join(strings.Fields(line), ""))
	}
	raw, err := hex.DecodeString(sb.String())
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return raw
}

func TestRoundTrip_EmptyPayload(t *testing.T) {
	raw := mustDecodeHexFile(t, "empty_payload.hex")
	n, f, err := ParseFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(raw) {
		t.Fatalf("n=%d want %d", n, len(raw))
	}
	if f.Version != VersionV1 || f.PayloadLen != 0 || len(f.Payload) != 0 {
		t.Fatalf("%+v", f)
	}
	out := AppendFrame(f)
	if !bytes.Equal(raw, out) {
		t.Fatalf("roundtrip\nwant %x\ngot  %x", raw, out)
	}
}

func TestRoundTrip_ThreeBytePayload(t *testing.T) {
	raw := mustDecodeHexFile(t, "three_byte_payload.hex")
	n, f, err := ParseFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(raw) {
		t.Fatalf("n=%d", n)
	}
	if string(f.Payload) != "Hel" {
		t.Fatalf("payload %q", f.Payload)
	}
	out := AppendFrame(f)
	if !bytes.Equal(raw, out) {
		t.Fatalf("roundtrip\nwant %x\ngot  %x", raw, out)
	}
}

func TestParseFrame_NeedMore_PartialHeader(t *testing.T) {
	// docs/spec/v1/frame-layout.md — fewer than 10 bytes available → ErrNeedMore (TRANS-01 半包).
	raw := mustDecodeHexFile(t, "need_more_short.hex")
	if len(raw) < 10 {
		t.Fatalf("fixture too short: %d", len(raw))
	}
	partial := raw[:5]
	_, _, err := ParseFrame(partial)
	if err != ErrNeedMore {
		t.Fatalf("want ErrNeedMore got %v", err)
	}
}

func TestParseFrame_ErrFrameTooLarge_FromFile(t *testing.T) {
	// docs/spec/v1/errors.md ERR_FRAME_TOO_LARGE; frame-layout.md payload_len 上限.
	raw := mustDecodeHexFile(t, "payload_too_large_prefix.hex")
	if len(raw) != HeaderSize {
		t.Fatalf("want %d-byte header got %d", HeaderSize, len(raw))
	}
	_, _, err := ParseFrame(raw)
	if err != ErrFrameTooLarge {
		t.Fatalf("want ErrFrameTooLarge got %v", err)
	}
}

func TestParseFrame_ErrProtoVersion_FromFile(t *testing.T) {
	// docs/spec/v1/errors.md ERR_PROTO_VERSION; version-capability 拒绝非 v1.
	raw := mustDecodeHexFile(t, "wrong_version.hex")
	_, _, err := ParseFrame(raw)
	if err != ErrProtoVersion {
		t.Fatalf("want ErrProtoVersion got %v", err)
	}
}

func TestErrFrameTooLarge(t *testing.T) {
	buf := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], MaxPayloadLen+1)
	_, _, err := ParseFrame(buf)
	if err != ErrFrameTooLarge {
		t.Fatalf("want ErrFrameTooLarge got %v", err)
	}
}

// TestGolden_RoutingModes_Broadcast — 见 docs/spec/v1/routing-modes.md「完整逻辑帧示例（广播）」与 testdata/framing/broadcast_stream_data_routing.hex。
func TestGolden_RoutingModes_Broadcast(t *testing.T) {
	raw := mustDecodeHexFile(t, "broadcast_stream_data_routing.hex")
	n, f, err := ParseFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(raw) {
		t.Fatalf("consumed %d want %d", n, len(raw))
	}
	if len(f.Payload) != 22 {
		t.Fatalf("payload len %d want 22", len(f.Payload))
	}
	if f.Payload[0] != 0x11 {
		t.Fatalf("msg_type %02x want STREAM_DATA 0x11", f.Payload[0])
	}
	if f.Payload[1] != 0x01 {
		t.Fatalf("routing_mode %02x want BROADCAST 0x01", f.Payload[1])
	}
	if binary.BigEndian.Uint64(f.Payload[2:10]) != 0xab {
		t.Fatalf("src_peer_id got %x want 0xab", binary.BigEndian.Uint64(f.Payload[2:10]))
	}
}
