package protocol

import (
	"strings"
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// RFC 6121 Example 2 (roster result; from 属性なしのサーバー応答)
const rfc6121Example2IQ = `<iq id='hf61v3n7' to='romeo@example.net/orchard' type='result'><query xmlns='jabber:iq:roster'><item jid='juliet@example.com' name='Juliet' subscription='both'><group>Friends</group></item></query></iq>`

func TestParseInboundIQ_RFC6121Example2(t *testing.T) {
	tok := iqToken(t, rfc6121Example2IQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	wantTo, _ := jid.Parse("romeo@example.net/orchard")
	if ev.To != wantTo {
		t.Fatalf("to = %v", ev.To)
	}
	if ev.ID != "hf61v3n7" || ev.Type != "result" {
		t.Fatalf("id=%q type=%q", ev.ID, ev.Type)
	}
	if !strings.Contains(string(ev.Payload), "jabber:iq:roster") {
		t.Fatalf("payload = %q", ev.Payload)
	}
	if !strings.Contains(string(ev.Payload), "juliet@example.com") {
		t.Fatal("missing roster item")
	}
}

func TestParseInboundIQ_Empty(t *testing.T) {
	tok := iqToken(t, `<iq id='x' type='get'/>`)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "x" || ev.Type != "get" || len(ev.Payload) != 0 {
		t.Fatalf("ev = %+v payload=%q", ev, ev.Payload)
	}
}

func TestParseInboundIQ_WithFrom(t *testing.T) {
	tok := iqToken(t, `<iq from='a@example.com/r' to='b@example.com' id='1' type='set'><ping xmlns='urn:xmpp:ping'/></iq>`)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	wantFrom, _ := jid.Parse("a@example.com/r")
	if ev.From != wantFrom {
		t.Fatalf("from = %v", ev.From)
	}
	if !strings.Contains(string(ev.Payload), "urn:xmpp:ping") {
		t.Fatalf("payload = %q", ev.Payload)
	}
}

func TestParseInboundIQ_InvalidFrom(t *testing.T) {
	tok := iqToken(t, `<iq from='@@' type='get'/>`)
	_, err := ParseInboundIQ(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInboundIQ_NotIQ(t *testing.T) {
	tok := iqToken(t, `<message/>`)
	_, err := ParseInboundIQ(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func iqToken(t *testing.T, xml string) xmlstream.Token {
	t.Helper()
	p := xmlstream.NewParser()
	p.Feed([]byte(xml))
	tok, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Kind != xmlstream.KindElement {
		t.Fatalf("kind = %v", tok.Kind)
	}
	return tok
}
