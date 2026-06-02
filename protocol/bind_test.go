package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func TestBindIQSetBytes_ServerGenerated(t *testing.T) {
	data, err := BindIQSetBytes("tn281v37", "")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !containsAll(s, "iq", "tn281v37", "set", "bind", bindNS) {
		t.Fatalf("bind iq: %s", s)
	}
	if contains(s, "<resource>") {
		t.Fatal("unexpected resource element")
	}
}

func TestBindIQSetBytes_ClientResource(t *testing.T) {
	data, err := BindIQSetBytes("wy2xa82b4", "balcony")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !containsAll(s, "wy2xa82b4", "balcony") {
		t.Fatalf("bind iq: %s", s)
	}
}

func TestParseBindResultJID_RFC6120_7_6_1(t *testing.T) {
	const raw = `<iq id='tn281v37' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
    <jid>juliet@im.example.com/4db06f06-1ea4-11dc-aca3-000bcd821bfb</jid>
  </bind>
</iq>`
	j, err := parseBindResultJID([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if j.Local() != "juliet" || j.Domain() != "im.example.com" || j.Resource() != "4db06f06-1ea4-11dc-aca3-000bcd821bfb" {
		t.Fatalf("jid = %s", j)
	}
}

func TestParseBindResultJID_RFC6120_7_7_1(t *testing.T) {
	const raw = `<iq id='wy2xa82b4' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
    <jid>juliet@im.example.com/balcony</jid>
  </bind>
</iq>`
	j, err := parseBindResultJID([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if j.Resource() != "balcony" {
		t.Fatalf("resource = %q", j.Resource())
	}
}

func TestBindIQSetBytes_EmptyID(t *testing.T) {
	if _, err := BindIQSetBytes("", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseBindResultJID_Missing(t *testing.T) {
	if _, err := parseBindResultJID([]byte(`<iq type='result'/>`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindResource_InvalidResource(t *testing.T) {
	j, _ := jid.Parse("u@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{BindOffered: true}
	if err := conn.BindResource("bad'res"); err == nil {
		t.Fatal("expected error")
	}
}
