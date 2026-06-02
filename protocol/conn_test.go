package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

const serverOpen = `<?xml version='1.0'?>
<stream:stream
    from='im.example.com'
    id='++TR84Sm6A3hnt3Q065SnAbbk3Y='
    to='juliet@im.example.com'
    version='1.0'
    xml:lang='en'
    xmlns='jabber:client'
    xmlns:stream='http://etherx.jabber.org/streams'>`

const featuresTLS = `<stream:features>
  <starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>
    <required/>
  </starttls>
</stream:features>`

func TestConnection_StreamOpenFlow(t *testing.T) {
	j, err := jid.Parse("juliet@im.example.com")
	if err != nil {
		t.Fatal(err)
	}
	conn := NewConnection(Config{JID: j, Lang: "en"})
	if err := conn.Start(); err != nil {
		t.Fatal(err)
	}
	if conn.State() != StateAwaitServerStream {
		t.Fatalf("state = %v", conn.State())
	}
	if conn.BytesToSend() == nil {
		t.Fatal("expected outbound stream open")
	}

	events, err := conn.Receive([]byte(serverOpen))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	opened, ok := events[0].(*StreamOpenedEvent)
	if !ok {
		t.Fatalf("event type %T", events[0])
	}
	if opened.ID != "++TR84Sm6A3hnt3Q065SnAbbk3Y=" {
		t.Fatalf("id = %q", opened.ID)
	}
	if conn.State() != StateAwaitFeatures {
		t.Fatalf("state = %v", conn.State())
	}

	events, err = conn.Receive([]byte(featuresTLS))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	if _, ok := events[0].(*StreamFeaturesEvent); !ok {
		t.Fatalf("event type %T", events[0])
	}
	if conn.State() != StateNegotiating {
		t.Fatalf("state = %v", conn.State())
	}
}

func TestConnection_NeedInput(t *testing.T) {
	j, _ := jid.Parse("juliet@im.example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()

	_, err := conn.Receive([]byte("<message><body>"))
	if err != nil {
		t.Fatal(err)
	}
	if !conn.NeedInput() {
		t.Fatal("NeedInput want true")
	}
}

func TestConnection_StanzaAfterReady(t *testing.T) {
	j, _ := jid.Parse("juliet@im.example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(`<stream:features/>`))
	conn.SetReady()

	const msg = `<message from='a@example.com' to='b@example.com'><body>hi</body></message>`
	events, err := conn.Receive([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	st, ok := events[0].(*StanzaEvent)
	if !ok || st.Name != "message" {
		t.Fatalf("event %T", events[0])
	}
}

func TestConnection_StartTwice(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	if err := conn.Start(); err != nil {
		t.Fatal(err)
	}
	if err := conn.Start(); err == nil {
		t.Fatal("expected error on second Start")
	}
}

func TestConnection_BytesToSendDrain(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	if conn.BytesToSend() != nil {
		t.Fatal("expected nil before Start")
	}
	_ = conn.Start()
	if conn.BytesToSend() == nil {
		t.Fatal("expected data after Start")
	}
	if conn.BytesToSend() != nil {
		t.Fatal("expected nil after drain")
	}
}

func TestConnection_StreamError(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))

	const streamErr = `<stream:error><not-well-formed xmlns='urn:ietf:params:xml:ns:xmpp-streams'/></stream:error>`
	events, err := conn.Receive([]byte(streamErr))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	if _, ok := events[0].(*StreamErrorEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	if conn.State() != StateClosed {
		t.Fatalf("state = %v", conn.State())
	}
	_, err = conn.Receive([]byte("<message/>"))
	if err == nil {
		t.Fatal("expected error after close")
	}
}

func TestConnection_StreamClose(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	conn.SetReady()

	events, err := conn.Receive([]byte("</stream:stream>"))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	if _, ok := events[0].(*StreamClosedEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
}

func TestConnection_StreamRestartInNegotiating(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	events, err := conn.Receive([]byte(serverOpen))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	if _, ok := events[0].(*StreamOpenedEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	if conn.State() != StateAwaitFeatures {
		t.Fatalf("state = %v", conn.State())
	}
}

func TestConnection_NegotiatingElementIgnored(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	events, err := conn.Receive([]byte(`<proceed xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("events = %d", len(events))
	}
}

func TestConnection_StanzaBeforeReady(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))

	_, err := conn.Receive([]byte(`<message><body>x</body></message>`))
	if err == nil {
		t.Fatal("expected error for stanza before ready")
	}
}

func TestConnection_UnexpectedStreamOpen(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_, err := conn.Receive([]byte(serverOpen))
	if err == nil {
		t.Fatal("expected error for stream open before Start")
	}
}

func TestConnection_FeaturesWrongState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_, err := conn.Receive([]byte(`<stream:features/>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SetReadyFromNegotiating(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(`<stream:features/>`))
	conn.SetReady()
	if conn.State() != StateReady {
		t.Fatalf("state = %v", conn.State())
	}
}

func TestConnection_StreamOpenInvalidFrom(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()

	const badFrom = `<?xml version='1.0'?><stream:stream from='@@' to='user@example.com' version='1.0' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>`
	_, err := conn.Receive([]byte(badFrom))
	if err == nil {
		t.Fatal("expected parse error for from")
	}
}

func TestConnection_StreamOpenEmptyFrom(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()

	const noFrom = `<?xml version='1.0'?><stream:stream to='user@example.com' id='x' version='1.0' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>`
	events, err := conn.Receive([]byte(noFrom))
	if err != nil {
		t.Fatal(err)
	}
	opened := events[0].(*StreamOpenedEvent)
	if !opened.From.IsEmpty() {
		t.Fatal("expected empty from")
	}
}

func TestConnection_StreamOpenWhenReady(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.SetReady()
	_, err := conn.Receive([]byte(serverOpen))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_UnknownElementWhenReady(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(`<stream:features/>`))
	conn.SetReady()

	events, err := conn.Receive([]byte(`<unknown xmlns='urn:example'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*ElementEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
}

func TestConnection_PresenceIQ(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(`<stream:features/>`))
	conn.SetReady()

	for _, name := range []string{"presence", "iq"} {
		events, err := conn.Receive([]byte("<" + name + "/>"))
		if err != nil {
			t.Fatal(err)
		}
		st, ok := events[0].(*StanzaEvent)
		if !ok || st.Name != name {
			t.Fatalf("name = %q", st.Name)
		}
	}
}
