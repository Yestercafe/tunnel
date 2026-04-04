// Package fakepeer provides a minimal TLS + v1 framing test harness for pkg/client
// integration tests. It is not the production Relay (Phase 9).
package fakepeer

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

type peerEnd struct {
	conn    *tls.Conn
	writeMu sync.Mutex
}

// Harness is a single-session fake with SESSION_CREATE / SESSION_JOIN and STREAM_DATA routing.
type Harness struct {
	mu sync.Mutex

	sessionID  string
	inviteCode string
	nextPeerID uint64

	serverTLS  *tls.Config
	clientPool *x509.CertPool

	peers map[uint64]*peerEnd
}

// NewHarness builds a harness with a valid session_id and invite_code for protocol.EncodeSessionCreateAck.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	srv, pool := LocalhostTLSConfig(t)
	return &Harness{
		sessionID:  randomSessionID(t),
		inviteCode: randomInviteCode(),
		nextPeerID: 1,
		serverTLS:  srv,
		clientPool: pool,
		peers:      make(map[uint64]*peerEnd),
	}
}

func randomSessionID(t *testing.T) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

func randomInviteCode() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	s := make([]byte, 8)
	for i := range s {
		s[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(s)
}

// SessionID returns the harness session id (valid for SESSION_CREATE_ACK).
func (h *Harness) SessionID() string { return h.sessionID }

// InviteCode returns the harness invite code.
func (h *Harness) InviteCode() string { return h.inviteCode }

// ClientTLS returns a tls.Config for dialing this harness (no InsecureSkipVerify).
func (h *Harness) ClientTLS() *tls.Config {
	return &tls.Config{
		RootCAs:    h.clientPool,
		ServerName: "127.0.0.1",
		MinVersion: tls.VersionTLS12,
	}
}

// Start listens on 127.0.0.1:0 and serves TLS. cleanup closes the listener and all tracked state.
func (h *Harness) Start(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr = tcpLn.Addr().String()
	ln := tls.NewListener(tcpLn, h.serverTLS)

	var wg sync.WaitGroup
	done := make(chan struct{})

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				select {
				case <-done:
				default:
				}
				return
			}
			tc, ok := c.(*tls.Conn)
			if !ok {
				c.Close()
				continue
			}
			wg.Add(1)
			go func(conn *tls.Conn) {
				defer wg.Done()
				h.handleConn(t, conn)
			}(tc)
		}
	}()

	return addr, func() {
		close(done)
		_ = ln.Close()
		wg.Wait()
	}
}

func (h *Harness) writeFrame(conn *tls.Conn, payload []byte) error {
	b := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})
	_, err := conn.Write(b)
	return err
}

func (h *Harness) handleConn(t *testing.T, conn *tls.Conn) {
	defer conn.Close()

	var peerID uint64
	var joined bool
	readBuf := make([]byte, 0, 8192)
	tmp := make([]byte, 4096)

	defer func() {
		if peerID != 0 {
			h.mu.Lock()
			delete(h.peers, peerID)
			h.mu.Unlock()
		}
	}()

	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			readBuf = append(readBuf, tmp[:n]...)
		}
		if err != nil {
			return
		}
		for {
			consumed, f, perr := framing.ParseFrame(readBuf)
			if perr == framing.ErrNeedMore {
				break
			}
			if perr != nil {
				return
			}
			readBuf = readBuf[consumed:]

			if len(f.Payload) == 0 {
				continue
			}
			switch f.Payload[0] {
			case protocol.MsgTypeSessionCreateReq:
				if err := protocol.DecodeSessionCreateReq(f.Payload); err != nil {
					return
				}
				ack, err := protocol.EncodeSessionCreateAck(h.sessionID, h.inviteCode)
				if err != nil {
					return
				}
				if err := h.writeFrame(conn, ack); err != nil {
					return
				}

			case protocol.MsgTypeSessionJoinReq:
				jb, cred, err := protocol.DecodeSessionJoinReq(f.Payload)
				if err != nil {
					return
				}
				ok := false
				switch jb {
				case 0:
					ok = cred == h.sessionID
				case 1:
					ok = cred == h.inviteCode
				}
				if !ok {
					pe, err := protocol.EncodeProtocolError(framing.ErrCodeJoinDenied, "credential mismatch")
					if err != nil {
						return
					}
					if err := h.writeFrame(conn, pe); err != nil {
						return
					}
					continue
				}

				h.mu.Lock()
				pid := h.nextPeerID
				h.nextPeerID++
				h.peers[pid] = &peerEnd{conn: conn}
				peerID = pid
				h.mu.Unlock()

				ack, err := protocol.EncodeSessionJoinAck(peerID)
				if err != nil {
					return
				}
				if err := h.writeFrame(conn, ack); err != nil {
					return
				}
				joined = true

			case protocol.MsgTypeStreamData:
				ok, _ := protocol.JoinGateAllowsBusinessDataPlane(joined, f.Payload)
				if !ok {
					pe, err := protocol.EncodeProtocolError(framing.ErrCodeJoinDenied, "data plane before join")
					if err != nil {
						return
					}
					if err := h.writeFrame(conn, pe); err != nil {
						return
					}
					continue
				}
				if peerID == 0 {
					return
				}
				h.forwardStreamData(peerID, f.Payload)

			case protocol.MsgTypeStreamOpen, protocol.MsgTypeStreamClose:
				ok, _ := protocol.JoinGateAllowsBusinessDataPlane(joined, f.Payload)
				if !ok {
					pe, err := protocol.EncodeProtocolError(framing.ErrCodeJoinDenied, "data plane before join")
					if err != nil {
						return
					}
					if err := h.writeFrame(conn, pe); err != nil {
						return
					}
					continue
				}
				// Harness does not simulate OPEN/CLOSE routing for Phase 8 fake; ignore or forward if needed later.

			default:
				// Ignore unknown for minimal harness.
			}
		}
	}
}

func (h *Harness) forwardStreamData(fromPeer uint64, payload []byte) {
	sd, err := protocol.DecodeStreamData(payload)
	if err != nil {
		return
	}
	if err := protocol.ValidateRoutingIntent(sd.Prefix.RoutingMode, sd.Prefix.DstPeerID); err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	switch sd.Prefix.RoutingMode {
	case protocol.RoutingModeBroadcast:
		for pid, pe := range h.peers {
			if pid == fromPeer {
				continue
			}
			h.writePeerFrame(pe, payload)
		}
	case protocol.RoutingModeUnicast:
		pe := h.peers[sd.Prefix.DstPeerID]
		if pe == nil || sd.Prefix.DstPeerID == fromPeer {
			return
		}
		h.writePeerFrame(pe, payload)
	}
}

func (h *Harness) writePeerFrame(pe *peerEnd, payload []byte) {
	pe.writeMu.Lock()
	defer pe.writeMu.Unlock()
	b := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})
	_, _ = pe.conn.Write(b)
}
