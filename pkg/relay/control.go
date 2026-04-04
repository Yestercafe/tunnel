package relay

import (
	"crypto/tls"
	"errors"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

func (s *Server) dispatchFrame(conn *tls.Conn, reg *SessionRegistry, st *connState, f framing.Frame) error {
	if len(f.Payload) == 0 {
		return nil
	}
	mt := f.Payload[0]
	switch mt {
	case protocol.MsgTypeSessionCreateReq:
		if err := protocol.DecodeSessionCreateReq(f.Payload); err != nil {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, err.Error())
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		sid, inv, err := reg.CreateSession()
		if err != nil {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, err.Error())
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		ack, err := protocol.EncodeSessionCreateAck(sid, inv)
		if err != nil {
			return err
		}
		return writeRawFrame(conn, ack)

	case protocol.MsgTypeSessionJoinReq:
		joinBy, cred, err := protocol.DecodeSessionJoinReq(f.Payload)
		if err != nil {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeJoinDenied, err.Error())
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		peerID, err := reg.JoinSession(joinBy, cred, conn)
		if err != nil {
			code := framing.ErrCodeSessionNotFound
			if errors.Is(err, ErrJoinDenied) {
				code = framing.ErrCodeJoinDenied
			}
			p, encErr := protocol.EncodeProtocolError(code, err.Error())
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		ack, err := protocol.EncodeSessionJoinAck(peerID)
		if err != nil {
			return err
		}
		st.joined = true
		st.peerID = peerID
		return writeRawFrame(conn, ack)

	case protocol.MsgTypeStreamOpen, protocol.MsgTypeStreamData, protocol.MsgTypeStreamClose:
		ok, err := protocol.JoinGateAllowsBusinessDataPlane(st.joined, f.Payload)
		if err != nil {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, err.Error())
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		if !ok {
			p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "ERR_ROUTING_INVALID")
			if encErr != nil {
				return encErr
			}
			return writeRawFrame(conn, p)
		}
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "relay: STREAM routing not implemented (phase 10)")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)

	default:
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "relay: unknown msg_type")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
}
