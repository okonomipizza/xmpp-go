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

const featuresSASL = `<stream:features>
  <mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>
    <mechanism>PLAIN</mechanism>
  </mechanisms>
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

func TestConnection_Features(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	f := conn.Features()
	if !f.StartTLSOffered || !f.StartTLSRequired {
		t.Fatalf("features = %+v", f)
	}
}

func TestConnection_StartTLSWrongState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	if err := conn.StartTLS(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_ProceedUnexpectedState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	_, err := conn.Receive([]byte(`<proceed xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>`))
	if err == nil {
		t.Fatal("expected error without StartTLS")
	}
}

func TestConnection_STARTTLSProceed(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	if err := conn.StartTLS(); err != nil {
		t.Fatal(err)
	}
	if string(conn.BytesToSend()) != string(StartTLSCommandBytes()) {
		t.Fatal("expected starttls command")
	}
	if conn.State() != StateAwaitTLSProceed {
		t.Fatalf("state = %s", conn.State())
	}

	events, err := conn.Receive([]byte(`<proceed xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	if _, ok := events[0].(*StartTLSProceedEvent); !ok {
		t.Fatalf("got %T", events[0])
	}

	conn.ResetAfterTLS()
	if conn.State() != StateInitial {
		t.Fatalf("state = %s", conn.State())
	}
	_ = conn.Start()
	if conn.State() != StateAwaitServerStream {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_STARTTLSFailure(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))
	_ = conn.StartTLS()
	_ = conn.BytesToSend()

	events, err := conn.Receive([]byte(`<failure xmlns='urn:ietf:params:xml:ns:xmpp-tls'/></stream:stream>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*StartTLSFailureEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	if conn.State() != StateClosed {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_StartTLSNotOffered(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(`<stream:features><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/></stream:features>`))

	if err := conn.StartTLS(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SASLFailure(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j, Password: "x"})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"PLAIN"}}
	_ = conn.Authenticate("PLAIN")
	_ = conn.BytesToSend()

	events, err := conn.Receive([]byte(`<failure xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*SASLFailureEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	if conn.State() != StateNegotiating {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_AuthenticateAutoSelectDefault(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j, Password: "p"})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1", "PLAIN"}}
	if err := conn.Authenticate(""); err != nil {
		t.Fatal(err)
	}
	if conn.State() != StateAwaitSASLOutcome {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_AuthenticateMechanismNotOffered(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j, Password: "p"})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"PLAIN"}}
	if err := conn.Authenticate("SCRAM-SHA-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_AuthenticateAutoSelect(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j, Password: "p", SASLMechanisms: []string{"PLAIN"}})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1", "PLAIN"}}
	if err := conn.Authenticate(""); err != nil {
		t.Fatal(err)
	}
	if conn.State() != StateAwaitSASLOutcome {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_SASLPLAINSuccess(t *testing.T) {
	j, _ := jid.Parse("juliet@im.example.com")
	conn := NewConnection(Config{
		JID:      j,
		Password: "r0m30myr0m30",
	})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresSASL))

	if err := conn.Authenticate("PLAIN"); err != nil {
		t.Fatal(err)
	}
	out := conn.BytesToSend()
	if out == nil {
		t.Fatal("expected auth element")
	}

	events, err := conn.Receive([]byte(`<success xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*SASLSuccessEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	conn.ResetAfterSASL()
	if conn.State() != StateInitial {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_SASLResponseWrongState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	if err := conn.SASLResponse([]byte("x")); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SASLChallengeResponse(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1"}}

	// SCRAM は内蔵 initial なし — 空 auth を送る用途は別途。challenge/response パスのみ検証
	data, err := AuthElementBytes("SCRAM-SHA-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn.enqueue(data)
	conn.state = StateAwaitSASLOutcome

	events, err := conn.Receive([]byte(`<challenge xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>Y2hhbGw=</challenge>`))
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := events[0].(*SASLChallengeEvent)
	if !ok || ch.Payload != "Y2hhbGw=" {
		t.Fatalf("challenge %+v", events[0])
	}

	if err := conn.SASLResponse([]byte("response")); err != nil {
		t.Fatal(err)
	}
	if conn.BytesToSend() == nil {
		t.Fatal("expected response")
	}
}

func TestConnection_SASLChallengeWrongState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	_, err := conn.Receive([]byte(`<challenge xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>eA==</challenge>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SASLChallengeWrongNamespace(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateAwaitSASLOutcome
	_, err := conn.Receive([]byte(`<challenge xmlns='urn:example'>eA==</challenge>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SASLSuccessWrongState(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	_, err := conn.Receive([]byte(`<success xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_BindServerGenerated(t *testing.T) {
	j, _ := jid.Parse("juliet@im.example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}

	if err := conn.Bind(); err != nil {
		t.Fatal(err)
	}
	if conn.State() != StateAwaitBind {
		t.Fatalf("state = %s", conn.State())
	}
	out := string(conn.BytesToSend())
	if !contains(out, "bind1") || !contains(out, "type='set'") {
		t.Fatalf("out = %s", out)
	}

	const result = `<iq id='bind1' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
    <jid>juliet@im.example.com/4db06f06-1ea4-11dc-aca3-000bcd821bfb</jid>
  </bind>
</iq>`
	events, err := conn.Receive([]byte(result))
	if err != nil {
		t.Fatal(err)
	}
	ev, ok := events[0].(*BindSuccessEvent)
	if !ok || ev.JID.Resource() != "4db06f06-1ea4-11dc-aca3-000bcd821bfb" {
		t.Fatalf("event = %+v", events[0])
	}
	if conn.State() != StateReady || conn.BoundJID().Resource() == "" {
		t.Fatalf("state = %s bound = %s", conn.State(), conn.BoundJID())
	}
}

func TestConnection_BindClientResource(t *testing.T) {
	j, _ := jid.Parse("juliet@im.example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}

	if err := conn.BindResource("balcony"); err != nil {
		t.Fatal(err)
	}
	_ = conn.BytesToSend()

	events, err := conn.Receive([]byte(`<iq id='bind1' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
    <jid>juliet@im.example.com/balcony</jid>
  </bind>
</iq>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*BindSuccessEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
}

func TestConnection_BindFailure(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}
	_ = conn.Bind()
	_ = conn.BytesToSend()

	events, err := conn.Receive([]byte(`<iq id='bind1' type='error'>
  <error type='cancel'><not-allowed xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'/></error>
</iq>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*BindFailureEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
	if conn.State() != StateNegotiating {
		t.Fatalf("state = %s", conn.State())
	}
}

func TestConnection_BindNotOffered(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	if err := conn.Bind(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_BindWrongIQType(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}
	_ = conn.Bind()
	_ = conn.BytesToSend()

	_, err := conn.Receive([]byte(`<iq id='bind1' type='set'/>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_IQAsStanzaWhenReady(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	events, err := conn.Receive([]byte(`<iq id='x' type='get'/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].(*StanzaEvent); !ok {
		t.Fatalf("got %T", events[0])
	}
}

func TestConnection_BindWrongIQID(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}
	_ = conn.Bind()
	_ = conn.BytesToSend()

	_, err := conn.Receive([]byte(`<iq id='other' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><jid>u@example.com/r</jid></bind>
</iq>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_NegotiatingElementIgnored(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	_ = conn.Start()
	_ = conn.BytesToSend()
	_, _ = conn.Receive([]byte(serverOpen))
	_, _ = conn.Receive([]byte(featuresTLS))

	events, err := conn.Receive([]byte(`<bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>`))
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
