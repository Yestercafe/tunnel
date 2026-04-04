package protocol

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadGoldenStreamDataMin(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "protocol", "stream_data_min.hex")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) != 1 {
		t.Fatalf("expected one hex line, got %q", lines)
	}
	raw := strings.ReplaceAll(lines[0], " ", "")
	out, err := hex.DecodeString(raw)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestStreamDataMinGolden(t *testing.T) {
	payload := loadGoldenStreamDataMin(t)
	if len(payload) != 25 {
		t.Fatalf("len %d", len(payload))
	}
	v, err := DecodeStreamData(payload)
	if err != nil {
		t.Fatal(err)
	}
	if v.Prefix.RoutingMode != RoutingModeBroadcast {
		t.Fatalf("mode %d", v.Prefix.RoutingMode)
	}
	if v.Prefix.SrcPeerID != 0xab {
		t.Fatalf("src %d", v.Prefix.SrcPeerID)
	}
	if v.Prefix.DstPeerID != 0 {
		t.Fatalf("dst %d", v.Prefix.DstPeerID)
	}
	if v.StreamID != 1 {
		t.Fatalf("stream_id %d", v.StreamID)
	}
	if v.InnerPayloadLen != 0 {
		t.Fatalf("inner %d", v.InnerPayloadLen)
	}
	prefix := RoutingPrefix{
		MsgType:     MsgTypeStreamData,
		RoutingMode: RoutingModeBroadcast,
		SrcPeerID:   0xab,
		DstPeerID:   0,
	}
	out, err := EncodeStreamData(prefix, 1, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(payload) {
		t.Fatalf("round-trip mismatch\n%x\n%x", out, payload)
	}
}

func TestStreamDataRoutingModesDocPrefix(t *testing.T) {
	// First 18 bytes match routing-modes.md broadcast example routing prefix (STREAM_DATA).
	payload := loadGoldenStreamDataMin(t)
	docPrefix, _ := hex.DecodeString("110100000000000000ab0000000000000000")
	if len(docPrefix) != 18 {
		t.Fatal(len(docPrefix))
	}
	if string(payload[:18]) != string(docPrefix) {
		t.Fatalf("prefix mismatch\n%x\n%x", payload[:18], docPrefix)
	}
}

func TestStreamCloseMin(t *testing.T) {
	pfx := RoutingPrefix{
		MsgType:     MsgTypeStreamClose,
		RoutingMode: RoutingModeBroadcast,
		SrcPeerID:   1,
		DstPeerID:   0,
	}
	p, err := EncodeStreamClose(pfx, 7)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeStreamClose(p)
	if err != nil {
		t.Fatal(err)
	}
	if v.StreamID != 7 || len(p) != 22 {
		t.Fatalf("sid=%d len=%d", v.StreamID, len(p))
	}
}

func TestRoutingPrefixErrors(t *testing.T) {
	_, err := ParseRoutingPrefix([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error")
	}
}
