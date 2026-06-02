package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
	"pgregory.net/rapid"
)

func validJID(t *rapid.T) jid.JID {
	s := rapid.StringMatching(`[a-z]{1,16}@[a-z0-9.-]{3,32}\.[a-z]{2,8}`).Draw(t, "jid")
	j, err := jid.Parse(s)
	if err != nil {
		t.Skip("generated JID invalid")
	}
	return j
}

func TestConnection_Property_StartState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		conn := NewConnection(Config{JID: validJID(t)})
		if err := conn.Start(); err != nil {
			t.Fatal(err)
		}
		if conn.State() != StateAwaitServerStream {
			t.Fatalf("state = %s", conn.State())
		}
	})
}

func TestConnection_Property_ServerOpenTransition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		conn := NewConnection(Config{JID: validJID(t)})
		_ = conn.Start()
		_, _ = conn.Receive([]byte(serverOpen))
		if conn.State() != StateAwaitFeatures {
			t.Fatalf("state = %s", conn.State())
		}
	})
}

func TestConnection_Property_StateIsKnown(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		conn := NewConnection(Config{JID: validJID(t)})
		chunk := rapid.SliceOfN(rapid.Byte(), 0, 256).Draw(t, "chunk")
		_ = conn.Start()
		_, _ = conn.Receive(chunk)

		switch conn.State() {
		case StateInitial, StateAwaitServerStream, StateAwaitFeatures,
			StateNegotiating, StateAwaitTLSProceed, StateAwaitSASLOutcome, StateReady, StateClosed:
		default:
			t.Fatalf("unknown state %s", conn.State())
		}
	})
}
