package protocol

import "testing"

func TestJoinGateMatrix(t *testing.T) {
	cases := []struct {
		name    string
		joined  bool
		msgType byte
		wantOK  bool
		wantErr bool
	}{
		{"empty", false, 0x00, false, true},
		{"ctrl create req not joined", false, 0x01, true, false},
		{"ctrl proto err not joined", false, 0x05, true, false},
		{"stream data blocked not joined", false, 0x11, false, false},
		{"stream open blocked not joined", false, 0x10, false, false},
		{"stream close blocked not joined", false, 0x12, false, false},
		{"stream data ok joined", true, 0x11, true, false},
		{"proto err not joined not data_plane_block", false, 0x05, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p []byte
			if tc.wantErr {
				p = []byte{}
			} else {
				p = []byte{tc.msgType}
			}
			ok, err := JoinGateAllowsBusinessDataPlane(tc.joined, p)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if ok != tc.wantOK {
				t.Fatalf("ok=%v want %v", ok, tc.wantOK)
			}
		})
	}
}

func TestJoinGateDistinguishesProtocolErrorFromStreamData(t *testing.T) {
	ok05, err := JoinGateAllowsBusinessDataPlane(false, []byte{0x05})
	if err != nil || !ok05 {
		t.Fatalf("PROTOCOL_ERROR should not be blocked as data plane: ok=%v err=%v", ok05, err)
	}
	ok11, err := JoinGateAllowsBusinessDataPlane(false, []byte{0x11})
	if err != nil || ok11 {
		t.Fatalf("STREAM_DATA must be blocked before join: ok=%v err=%v", ok11, err)
	}
}
