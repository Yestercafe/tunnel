package relay

import (
	"crypto/tls"
	"errors"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

// connState tracks per-connection session join state (RLY-02).
type connState struct {
	joined bool
	peerID uint64
}

func (s *Server) serveConn(conn *tls.Conn) {
	defer conn.Close()

	reg := s.Registry
	if reg == nil {
		return
	}

	var st connState
	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			return
		}
		for len(buf) > 0 {
			consumed, f, err := framing.ParseFrame(buf)
			if errors.Is(err, framing.ErrNeedMore) {
				break
			}
			if err != nil {
				if werr := s.writeFramingError(conn, err); werr != nil {
					return
				}
				return
			}
			buf = buf[consumed:]
			if err := s.dispatchFrame(conn, reg, &st, f); err != nil {
				return
			}
		}
	}
}

func (s *Server) writeFramingError(conn *tls.Conn, ferr error) error {
	var code framing.ErrCode
	switch {
	case errors.Is(ferr, framing.ErrFrameTooLarge):
		code = framing.ErrCodeFrameTooLarge
	case errors.Is(ferr, framing.ErrProtoVersion):
		code = framing.ErrCodeProtoVersion
	default:
		code = framing.ErrCodeProtoVersion
	}
	reason := ferr.Error()
	p, err := protocol.EncodeProtocolError(code, reason)
	if err != nil {
		return err
	}
	return writeRawFrame(conn, p)
}

func writeRawFrame(conn *tls.Conn, payload []byte) error {
	wire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})
	_, err := conn.Write(wire)
	return err
}
