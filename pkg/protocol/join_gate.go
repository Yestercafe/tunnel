package protocol

import "errors"

// JoinGateAllowsBusinessDataPlane enforces session-state.md STATE-01 for data-plane opcodes.
//
// JoinGateAllowsBusinessDataPlane reports whether the payload may proceed as data-plane stream
// business traffic. When joined is false, STREAM_OPEN/DATA/CLOSE (0x10–0x12) are not allowed;
// PROTOCOL_ERROR (0x05) and SESSION_* control messages are not blocked by this gate.
// Empty payload returns an error.
func JoinGateAllowsBusinessDataPlane(joined bool, payload []byte) (ok bool, err error) {
	if len(payload) == 0 {
		return false, errors.New("protocol: empty payload")
	}
	mt := payload[0]
	if !joined {
		switch mt {
		case MsgTypeStreamOpen, MsgTypeStreamData, MsgTypeStreamClose:
			return false, nil
		default:
			return true, nil
		}
	}
	return true, nil
}
