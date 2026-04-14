package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

// Client is a synchronous Tunnel client over an established TLS connection.
type Client struct {
	conn   *tls.Conn
	readBuf []byte

	joined bool
	peerID uint64
}

// Dial opens a TCP connection to addr, performs TLS, and returns a Client.
func Dial(ctx context.Context, addr string, tlsCfg *tls.Config) (*Client, error) {
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	cfg := tlsCfg.Clone()

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("client: dial: %w", err)
	}

	if cfg.ServerName == "" {
		if ip := net.ParseIP(host); ip == nil {
			cfg.ServerName = host
		}
	}

	d := &net.Dialer{}
	raw, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	tc := tls.Client(raw, cfg)
	if err := tc.HandshakeContext(ctx); err != nil {
		raw.Close()
		return nil, err
	}
	return &Client{conn: tc}, nil
}

// Close closes the TLS connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) writeFrame(ctx context.Context, payload []byte) error {
	if d, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(d)
	} else {
		_ = c.conn.SetWriteDeadline(time.Time{})
	}
	wire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})
	_, err := c.conn.Write(wire)
	return err
}

func (c *Client) readFrame(ctx context.Context) (framing.Frame, error) {
	tmp := make([]byte, 4096)
	for {
		consumed, f, err := framing.ParseFrame(c.readBuf)
		if err == nil {
			c.readBuf = c.readBuf[consumed:]
			if err := c.gateInboundDataPlane(f.Payload); err != nil {
				return framing.Frame{}, err
			}
			return f, nil
		}
		if !errors.Is(err, framing.ErrNeedMore) {
			return framing.Frame{}, err
		}
		if d, ok := ctx.Deadline(); ok {
			_ = c.conn.SetReadDeadline(d)
		} else {
			_ = c.conn.SetReadDeadline(time.Time{})
		}
		n, rerr := c.conn.Read(tmp)
		if n > 0 {
			c.readBuf = append(c.readBuf, tmp[:n]...)
		}
		if rerr != nil {
			return framing.Frame{}, rerr
		}
	}
}

func (c *Client) gateInboundDataPlane(payload []byte) error {
	if len(payload) == 0 {
		return nil
	}
	mt := payload[0]
	switch mt {
	case protocol.MsgTypeStreamOpen, protocol.MsgTypeStreamData, protocol.MsgTypeStreamClose:
		ok, err := protocol.JoinGateAllowsBusinessDataPlane(c.joined, payload)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNotJoined
		}
	}
	return nil
}

// CreateSession sends SESSION_CREATE_REQ and waits for SESSION_CREATE_ACK.
func (c *Client) CreateSession(ctx context.Context) (sessionID string, inviteCode string, err error) {
	if err := c.writeFrame(ctx, protocol.EncodeSessionCreateReq()); err != nil {
		return "", "", err
	}
	for {
		f, err := c.readFrame(ctx)
		if err != nil {
			return "", "", err
		}
		if len(f.Payload) == 0 {
			continue
		}
		switch f.Payload[0] {
		case protocol.MsgTypeSessionCreateAck:
			return protocol.DecodeSessionCreateAck(f.Payload)
		case protocol.MsgTypeProtocolError:
			code, reason, err := protocol.DecodeProtocolError(f.Payload)
			if err != nil {
				return "", "", err
			}
			return "", "", &ProtocolError{Code: code, Reason: reason}
		default:
			continue
		}
	}
}

// JoinSession sends SESSION_JOIN_REQ. If bySessionID is true, credential is session_id; otherwise invite_code.
func (c *Client) JoinSession(ctx context.Context, bySessionID bool, credential string) (peerID uint64, err error) {
	var joinBy uint8 = 1
	if bySessionID {
		joinBy = 0
	}
	payload, err := protocol.EncodeSessionJoinReq(joinBy, credential)
	if err != nil {
		return 0, err
	}
	if err := c.writeFrame(ctx, payload); err != nil {
		return 0, err
	}
	for {
		f, err := c.readFrame(ctx)
		if err != nil {
			return 0, err
		}
		if len(f.Payload) == 0 {
			continue
		}
		switch f.Payload[0] {
		case protocol.MsgTypeSessionJoinAck:
			pid, err := protocol.DecodeSessionJoinAck(f.Payload)
			if err != nil {
				return 0, err
			}
			c.peerID = pid
			c.joined = true
			return pid, nil
		case protocol.MsgTypeProtocolError:
			code, reason, err := protocol.DecodeProtocolError(f.Payload)
			if err != nil {
				return 0, err
			}
			return 0, &ProtocolError{Code: code, Reason: reason}
		default:
			continue
		}
	}
}

// SendStreamData sends STREAM_DATA with routing prefix + stream fields + application data.
func (c *Client) SendStreamData(ctx context.Context, mode uint8, dstPeerID uint64, streamID uint32, flags uint8, app []byte) error {
	prefix := protocol.RoutingPrefix{
		MsgType:     protocol.MsgTypeStreamData,
		RoutingMode: mode,
		SrcPeerID:   c.peerID,
		DstPeerID:   dstPeerID,
	}
	if err := protocol.ValidateRoutingIntent(mode, dstPeerID); err != nil {
		return err
	}
	payload, err := protocol.EncodeStreamData(prefix, streamID, flags, app)
	if err != nil {
		return err
	}
	ok, err := protocol.JoinGateAllowsBusinessDataPlane(c.joined, payload)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotJoined
	}
	return c.writeFrame(ctx, payload)
}

// ReadFrame returns the next full frame from the wire (after inbound join-gate checks).
func (c *Client) ReadFrame(ctx context.Context) (framing.Frame, error) {
	return c.readFrame(ctx)
}

// UnderlyingTLSConn returns the established TLS connection. Intended for integration tests that must send frames not exposed by the public API (e.g. STREAM_DATA before JOIN per Phase 11).
func (c *Client) UnderlyingTLSConn() *tls.Conn {
	if c == nil {
		return nil
	}
	return c.conn
}

// PeerID returns the local peer id after a successful JoinSession.
func (c *Client) PeerID() uint64 { return c.peerID }
