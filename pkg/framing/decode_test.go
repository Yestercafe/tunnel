package framing

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestRoundTrip_EmptyPayload(t *testing.T) {
	// docs/spec/v1/frame-layout.md example: no payload, version 0x0001, cap 0
	h := "00000000000100000000"
	raw, err := hex.DecodeString(h)
	if err != nil {
		t.Fatal(err)
	}
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
	h := "0000000300010000000048656c"
	raw, err := hex.DecodeString(h)
	if err != nil {
		t.Fatal(err)
	}
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

func TestErrFrameTooLarge(t *testing.T) {
	buf := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], MaxPayloadLen+1)
	_, _, err := ParseFrame(buf)
	if err != ErrFrameTooLarge {
		t.Fatalf("want ErrFrameTooLarge got %v", err)
	}
}
