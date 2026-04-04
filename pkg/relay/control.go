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
		peerID, sessionID, err := reg.JoinSession(joinBy, cred, conn)
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
		st.sessionID = sessionID
		return writeRawFrame(conn, ack)

	case protocol.MsgTypeStreamOpen, protocol.MsgTypeStreamClose:
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
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "relay: STREAM_OPEN/STREAM_CLOSE not implemented (phase 10)")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)

	case protocol.MsgTypeStreamData:
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
		return s.dispatchStreamData(conn, reg, st, f)

	default:
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "relay: unknown msg_type")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
}

func (s *Server) dispatchStreamData(conn *tls.Conn, reg *SessionRegistry, st *connState, f framing.Frame) error {
	if !st.joined || st.sessionID == "" {
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "relay: not joined")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
	v, err := protocol.DecodeStreamData(f.Payload)
	if err != nil {
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, err.Error())
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
	if v.Prefix.SrcPeerID != st.peerID {
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, "src_peer_id mismatch")
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
	if err := protocol.ValidateRoutingIntent(v.Prefix.RoutingMode, v.Prefix.DstPeerID); err != nil {
		p, encErr := protocol.EncodeProtocolError(framing.ErrCodeRoutingInvalid, err.Error())
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
	if err := reg.DeliverStreamData(st.sessionID, st.peerID, v.Prefix.RoutingMode, v.Prefix.DstPeerID, f.Payload); err != nil {
		var code framing.ErrCode
		switch {
		case errors.Is(err, ErrSessionNotFound):
			code = framing.ErrCodeSessionNotFound
		case errors.Is(err, ErrDstPeerNotInSession):
			code = framing.ErrCodeRoutingInvalid
		default:
			return err
		}
		p, encErr := protocol.EncodeProtocolError(code, err.Error())
		if encErr != nil {
			return encErr
		}
		return writeRawFrame(conn, p)
	}
	return nil
}
