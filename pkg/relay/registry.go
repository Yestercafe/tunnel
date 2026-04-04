package relay

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"

	"tunnel/pkg/framing"
	"tunnel/pkg/protocol"
)

// FrameWriter is a minimal write surface for outbound frames (typically *tls.Conn).
type FrameWriter interface {
	Write(p []byte) (n int, err error)
}

var (
	// ErrSessionNotFound is returned when JOIN references an unknown session or invite.
	ErrSessionNotFound = errors.New("relay: session not found")
	// ErrJoinDenied is returned when JOIN credentials are invalid.
	ErrJoinDenied = errors.New("relay: join denied")
	// ErrDstPeerNotInSession is returned when unicast targets a missing peer_id.
	ErrDstPeerNotInSession = errors.New("relay: dst peer not in session")
)

// Peer is a joined member of a session (RLY-03 will use this for routing).
type Peer struct {
	ID uint64
	W  FrameWriter
}

type session struct {
	id         string
	inviteCode string
	nextPeerID uint64
	peers      map[uint64]*Peer
}

// SessionRegistry is an in-memory session → peer map (single process).
type SessionRegistry struct {
	mu       sync.Mutex
	byID     map[string]*session
	byInvite map[string]*session
}

// NewSessionRegistry returns an empty registry.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		byID:     make(map[string]*session),
		byInvite: make(map[string]*session),
	}
}

// CreateSession allocates a new session_id (UUID v4) and invite_code (8–12 base32 chars).
func (r *SessionRegistry) CreateSession() (sessionID, inviteCode string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	uid, err := uuid.NewRandom()
	if err != nil {
		return "", "", err
	}
	sessionID = uid.String()

	for attempts := 0; attempts < 10; attempts++ {
		inviteCode, err = randomInviteCode()
		if err != nil {
			return "", "", err
		}
		if _, exists := r.byInvite[inviteCode]; exists {
			continue
		}
		if _, err := protocol.EncodeSessionCreateAck(sessionID, inviteCode); err != nil {
			return "", "", err
		}
		sess := &session{
			id:         sessionID,
			inviteCode: inviteCode,
			peers:      make(map[uint64]*Peer),
		}
		r.byID[sessionID] = sess
		r.byInvite[inviteCode] = sess
		return sessionID, inviteCode, nil
	}
	return "", "", errors.New("relay: could not allocate unique invite code")
}

// JoinSession registers a new peer in the session identified by credential.
// joinBy 0 = session_id, 1 = invite_code.
func (r *SessionRegistry) JoinSession(joinBy uint8, credential string, w FrameWriter) (peerID uint64, sessionID string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var sess *session
	switch joinBy {
	case 0:
		sess = r.byID[credential]
	case 1:
		sess = r.byInvite[credential]
	default:
		return 0, "", fmt.Errorf("%w: invalid join_by", ErrJoinDenied)
	}
	if sess == nil {
		return 0, "", ErrSessionNotFound
	}

	sess.nextPeerID++
	peerID = sess.nextPeerID
	if peerID == 0 {
		return 0, "", errors.New("relay: peer_id overflow")
	}
	sess.peers[peerID] = &Peer{ID: peerID, W: w}
	return peerID, sess.id, nil
}

// DeliverStreamData forwards a full STREAM_DATA payload to peers in sessionID.
// Broadcast sends to all peers except senderPeerID (no echo to sender).
// Unicast sends only to dstPeerID when mode is RoutingModeUnicast.
func (r *SessionRegistry) DeliverStreamData(sessionID string, senderPeerID uint64, mode uint8, dstPeerID uint64, payload []byte) error {
	wire := framing.AppendFrame(framing.Frame{
		Version:    framing.VersionV1,
		Capability: 0,
		Payload:    payload,
	})

	r.mu.Lock()
	sess := r.byID[sessionID]
	if sess == nil {
		r.mu.Unlock()
		return ErrSessionNotFound
	}

	var targets []*Peer
	switch mode {
	case protocol.RoutingModeBroadcast:
		for id, p := range sess.peers {
			if id == senderPeerID {
				continue
			}
			targets = append(targets, p)
		}
	case protocol.RoutingModeUnicast:
		p := sess.peers[dstPeerID]
		if p == nil {
			r.mu.Unlock()
			return ErrDstPeerNotInSession
		}
		targets = []*Peer{p}
	default:
		r.mu.Unlock()
		return fmt.Errorf("relay: unknown routing_mode %d", mode)
	}
	r.mu.Unlock()

	for _, p := range targets {
		if _, err := p.W.Write(wire); err != nil {
			return err
		}
	}
	return nil
}

func randomInviteCode() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	n := 8 + int(randByte()%5) // 8–12 inclusive
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		if err := b.WriteByte(alphabet[randByte()%byte(len(alphabet))]); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

func randByte() byte {
	var b [1]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return 0
	}
	return b[0]
}
