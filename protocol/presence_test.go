package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// RFC 6121 Example 11 (サーバーから contact へ broadcast する presence の 1 件)
const rfc6121Example11Presence = `<presence from='romeo@example.net/orchard' to='juliet@example.com' xml:lang='en'><show>away</show><status>I shall return!</status><priority>1</priority></presence>`

func TestParseInboundPresence_RFC6121Example11(t *testing.T) {
	tok := presenceToken(t, rfc6121Example11Presence)
	ev, err := ParseInboundPresence(tok)
	if err != nil {
		t.Fatal(err)
	}
	wantFrom, _ := jid.Parse("romeo@example.net/orchard")
	wantTo, _ := jid.Parse("juliet@example.com")
	if ev.From != wantFrom {
		t.Fatalf("from = %v", ev.From)
	}
	if ev.To != wantTo {
		t.Fatalf("to = %v", ev.To)
	}
	if ev.Lang != "en" {
		t.Fatalf("lang = %q", ev.Lang)
	}
	if ev.Show != "away" {
		t.Fatalf("show = %q", ev.Show)
	}
	if ev.Status != "I shall return!" {
		t.Fatalf("status = %q", ev.Status)
	}
	if ev.Priority == nil || *ev.Priority != 1 {
		t.Fatalf("priority = %v", ev.Priority)
	}
}

func TestParseInboundPresence_Probe(t *testing.T) {
	tok := presenceToken(t, `<presence from='a@example.com' to='b@example.com' type='probe'/>`)
	ev, err := ParseInboundPresence(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "probe" {
		t.Fatalf("type = %q", ev.Type)
	}
	if ev.Show != "" || ev.Priority != nil {
		t.Fatalf("unexpected children show=%q priority=%v", ev.Show, ev.Priority)
	}
}

func TestParseInboundPresence_UnescapeStatus(t *testing.T) {
	tok := presenceToken(t, `<presence from='a@example.com'><status>Tom &amp; Jerry</status></presence>`)
	ev, err := ParseInboundPresence(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Status != "Tom & Jerry" {
		t.Fatalf("status = %q", ev.Status)
	}
}

func TestParseInboundPresence_InvalidPriority(t *testing.T) {
	tok := presenceToken(t, `<presence from='a@example.com'><priority>x</priority></presence>`)
	_, err := ParseInboundPresence(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInboundPresence_InvalidFrom(t *testing.T) {
	tok := presenceToken(t, `<presence from='@@'/>`)
	_, err := ParseInboundPresence(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInboundPresence_NotPresence(t *testing.T) {
	tok := presenceToken(t, `<message/>`)
	_, err := ParseInboundPresence(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func presenceToken(t *testing.T, xml string) xmlstream.Token {
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
