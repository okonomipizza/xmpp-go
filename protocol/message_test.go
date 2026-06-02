package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// RFC 6121 Example 9 (最初の message)
const rfc6121Example9Message1 = `<message from='juliet@example.com/balcony' to='romeo@example.net' type='chat' xml:lang='en'><body>My ears have not yet drunk a hundred words</body><thread>e0ffe42b28561960c6b12b944a092794b9683a38</thread></message>`

func TestParseInboundMessage_RFC6121Example9(t *testing.T) {
	tok := messageToken(t, rfc6121Example9Message1)
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	wantFrom, _ := jid.Parse("juliet@example.com/balcony")
	wantTo, _ := jid.Parse("romeo@example.net")
	if ev.From != wantFrom {
		t.Fatalf("from = %v", ev.From)
	}
	if ev.To != wantTo {
		t.Fatalf("to = %v", ev.To)
	}
	if ev.Type != "chat" {
		t.Fatalf("type = %q", ev.Type)
	}
	if ev.Lang != "en" {
		t.Fatalf("lang = %q", ev.Lang)
	}
	if ev.Body != "My ears have not yet drunk a hundred words" {
		t.Fatalf("body = %q", ev.Body)
	}
}

func TestParseInboundMessage_NoBody(t *testing.T) {
	tok := messageToken(t, `<message from='a@example.com' to='b@example.com'/>`)
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Body != "" {
		t.Fatalf("body = %q", ev.Body)
	}
}

func TestParseInboundMessage_FirstBodyWins(t *testing.T) {
	tok := messageToken(t, `<message from='a@example.com' to='b@example.com'><body>one</body><body>two</body></message>`)
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Body != "one" {
		t.Fatalf("body = %q", ev.Body)
	}
}

func TestParseInboundMessage_UnescapeBody(t *testing.T) {
	tok := messageToken(t, `<message from='a@example.com' to='b@example.com'><body>Tom &amp; Jerry</body></message>`)
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Body != "Tom & Jerry" {
		t.Fatalf("body = %q", ev.Body)
	}
}

func TestParseInboundMessage_UnescapeLtGt(t *testing.T) {
	tok := messageToken(t, `<message from='a@example.com' to='b@example.com'><body>1 &lt; 2 &gt; 0</body></message>`)
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Body != "1 < 2 > 0" {
		t.Fatalf("body = %q", ev.Body)
	}
}

func TestParseInboundMessage_RoundtripWithMessageStanzaBytes(t *testing.T) {
	from, _ := jid.Parse("a@example.com/res")
	to, _ := jid.Parse("b@example.com")
	const body = "Tom & Jerry <end>"
	data, err := MessageStanzaBytes(from, to, "1", "chat", "", body)
	if err != nil {
		t.Fatal(err)
	}
	tok := messageToken(t, string(data))
	ev, err := ParseInboundMessage(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Body != body {
		t.Fatalf("body = %q want %q", ev.Body, body)
	}
}

func TestParseInboundMessage_InvalidFrom(t *testing.T) {
	tok := messageToken(t, `<message from='@@' to='b@example.com'><body>x</body></message>`)
	_, err := ParseInboundMessage(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInboundMessage_NotMessage(t *testing.T) {
	tok := messageToken(t, `<presence/>`)
	_, err := ParseInboundMessage(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func messageToken(t *testing.T, xml string) xmlstream.Token {
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
