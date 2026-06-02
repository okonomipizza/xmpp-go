package protocol

import (
	"strings"
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func TestIQStanzaBytes_EmptyInner(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	data, err := IQStanzaBytes(from, jid.JID{}, "1", "get", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "/>") {
		t.Fatalf("stanza: %s", data)
	}
}

func TestMessageStanzaBytes_WithLang(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	data, err := MessageStanzaBytes(from, to, "1", "chat", "en", "hi")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "xml:lang='en'") {
		t.Fatalf("stanza: %s", data)
	}
}

func TestPresenceStanzaBytes_AllAttrs(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	data, err := PresenceStanzaBytes(from, to, "p1", "subscribe")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !containsAll(s, "to='v@example.com'", "id='p1'", "type='subscribe'") {
		t.Fatalf("stanza: %s", s)
	}
}

func TestMessageStanzaBytes_InvalidLang(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	if _, err := MessageStanzaBytes(from, to, "1", "chat", "en>", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIQStanzaBytes_AllTypes(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	for _, typ := range []string{"get", "set", "result", "error"} {
		if _, err := IQStanzaBytes(from, jid.JID{}, "1", typ, nil); err != nil {
			t.Fatalf("type %s: %v", typ, err)
		}
	}
}

func TestMessageStanzaBytes_EmptyID(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	if _, err := MessageStanzaBytes(from, to, "", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendMessageInvalidType(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendMessage(to, "1", "", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestMessageStanzaBytes_RFC6121_5_2_1(t *testing.T) {
	from, _ := jid.Parse("juliet@example.com/balcony")
	to, _ := jid.Parse("romeo@example.net")
	data, err := MessageStanzaBytes(from, to, "ktx72v49", "chat", "en",
		"Art thou not Romeo, and a Montague?")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "juliet@example.com/balcony") ||
		!strings.Contains(s, "romeo@example.net") ||
		!strings.Contains(s, "Art thou not Romeo") {
		t.Fatalf("stanza: %s", s)
	}
}

func TestPresenceStanzaBytes_IDOnly(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	data, err := PresenceStanzaBytes(from, jid.JID{}, "p1", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "id='p1'") || strings.Contains(string(data), "type=") {
		t.Fatalf("stanza: %s", data)
	}
}

func TestPresenceStanzaBytes_InvalidType(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	if _, err := PresenceStanzaBytes(from, jid.JID{}, "", "bad<"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIQStanzaBytes_WithTo(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	if _, err := IQStanzaBytes(from, to, "1", "get", nil); err != nil {
		t.Fatal(err)
	}
}

func TestPresenceStanzaBytes_NoTo(t *testing.T) {
	from, _ := jid.Parse("user@example.com/res")
	data, err := PresenceStanzaBytes(from, jid.JID{}, "", "unavailable")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), " to=") {
		t.Fatalf("stanza: %s", data)
	}
}

func TestIQStanzaBytes_WithInner(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("example.com")
	inner := []byte(`<query xmlns='jabber:iq:version'/>`)
	data, err := IQStanzaBytes(from, to, "q1", "get", inner)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "<query") || !strings.HasSuffix(s, "</iq>") {
		t.Fatalf("stanza: %s", s)
	}
}

func TestMessageStanzaBytes_EmptyTo(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	if _, err := MessageStanzaBytes(from, jid.JID{}, "1", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestMessageStanzaBytes_BareFrom(t *testing.T) {
	from, _ := jid.Parse("user@example.com")
	to, _ := jid.Parse("other@example.com")
	if _, err := MessageStanzaBytes(from, to, "1", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendMessageNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com")
	conn := NewConnection(Config{JID: j})
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendMessage(to, "1", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendPresenceNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendPresence(to, "p1", "subscribe"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendIQNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendIQ(to, "q1", "get", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendPresenceInvalidType(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendPresence(to, "p1", "bad<"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendIQInvalidType(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendIQ(to, "q1", "invalid", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestIQStanzaBytes_InvalidType(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	if _, err := IQStanzaBytes(from, jid.JID{}, "1", "invalid", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestMessageStanzaBytes_ForbiddenID(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	if _, err := MessageStanzaBytes(from, to, "id>", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestEscapeXMLText(t *testing.T) {
	if got := escapeXMLText("Tom & Jerry"); got != "Tom &amp; Jerry" {
		t.Fatalf("got %q", got)
	}
	if got := escapeXMLText("x < y"); got != "x &lt; y" {
		t.Fatalf("got %q", got)
	}
	if got := escapeXMLText("a > b"); got != "a &gt; b" {
		t.Fatalf("got %q", got)
	}
}

func TestMessageStanzaBytes_EscapesBody(t *testing.T) {
	from, _ := jid.Parse("u@example.com/r")
	to, _ := jid.Parse("v@example.com")
	data, err := MessageStanzaBytes(from, to, "1", "chat", "", "Tom & Jerry <3")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "Tom &amp; Jerry &lt;3") {
		t.Fatalf("stanza: %s", s)
	}
}

func TestConnection_SendWithoutBoundJID(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendMessage(to, "1", "chat", "", "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendPresenceAndIQ(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendPresence(to, "p1", "subscribe"); err != nil {
		t.Fatal(err)
	}
	_ = conn.BytesToSend()
	if err := conn.SendIQ(to, "q1", "get", []byte("<ping/>")); err != nil {
		t.Fatal(err)
	}
	if conn.BytesToSend() == nil {
		t.Fatal("expected iq")
	}
}

func TestConnection_SendMessageAfterBind(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	to, _ := jid.Parse("v@example.com")
	if err := conn.SendMessage(to, "1", "chat", "", "hi"); err != nil {
		t.Fatal(err)
	}
	if conn.BytesToSend() == nil {
		t.Fatal("expected outbound stanza")
	}
}
