package protocol

import (
	"strings"
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func TestRosterIQGetBytes_RFC6121Example1(t *testing.T) {
	from, _ := jid.Parse("romeo@example.net/orchard")
	data, err := RosterIQGetBytes(from, "hf61v3n7")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		"from='romeo@example.net/orchard'",
		"id='hf61v3n7'",
		"type='get'",
		"<query xmlns='jabber:iq:roster'/>",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestParseRosterItems_FromIQEvent(t *testing.T) {
	tok := iqToken(t, rfc6121Example2IQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	items, err := ParseRosterItems(ev.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestParseRosterItems_RFC6121Example2(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'>` +
		`<item jid='juliet@example.com' name='Juliet' subscription='both'><group>Friends</group></item>` +
		`<item jid='benvolio@example.org' name='Benvolio' subscription='to'/>` +
		`<item jid='mercutio@example.org' name='Mercutio' subscription='from'/>` +
		`</query>`)
	items, err := ParseRosterItems(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("items = %d", len(items))
	}
	j0, _ := jid.Parse("juliet@example.com")
	if items[0].JID != j0 || items[0].Name != "Juliet" || items[0].Subscription != "both" {
		t.Fatalf("item0 = %+v", items[0])
	}
	if len(items[0].Groups) != 1 || items[0].Groups[0] != "Friends" {
		t.Fatalf("groups = %v", items[0].Groups)
	}
	j1, _ := jid.Parse("benvolio@example.org")
	if items[1].JID != j1 || items[1].Subscription != "to" {
		t.Fatalf("item1 = %+v", items[1])
	}
}

func TestParseRosterItems_Empty(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'/>`)
	items, err := ParseRosterItems(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %v", items)
	}
}

func TestParseRosterItems_NoQuery(t *testing.T) {
	_, err := ParseRosterItems([]byte(`<ping xmlns='urn:xmpp:ping'/>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterItems_InvalidJID(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'><item jid='@@'/></query>`)
	_, err := ParseRosterItems(payload)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterItems_QueryAsChild(t *testing.T) {
	payload := []byte(`<wrapper><query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query></wrapper>`)
	items, err := ParseRosterItems(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestParseRosterItems_WrongNamespace(t *testing.T) {
	_, err := ParseRosterItems([]byte(`<query xmlns='urn:example:other'/>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendRosterGetNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.boundJID = j
	if err := conn.SendRosterGet("r1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendRosterGet(t *testing.T) {
	j, _ := jid.Parse("romeo@example.net/orchard")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	if err := conn.SendRosterGet("r1"); err != nil {
		t.Fatal(err)
	}
	data := conn.BytesToSend()
	if data == nil || !strings.Contains(string(data), "jabber:iq:roster") {
		t.Fatalf("data: %s", data)
	}
}

func TestElementAttrValue(t *testing.T) {
	single := []byte(`<item jid='a@example.com' name='A' subscription='both'/>`)
	if got := elementAttrValue(single, "jid"); got != "a@example.com" {
		t.Fatalf("jid = %q", got)
	}
	if got := elementAttrValue(single, "subscription"); got != "both" {
		t.Fatalf("subscription = %q", got)
	}
	double := []byte(`<item jid="a@example.com" subscription="both"/>`)
	if got := elementAttrValue(double, "jid"); got != "a@example.com" {
		t.Fatalf("jid = %q", got)
	}
	if got := elementAttrValue(double, "subscription"); got != "both" {
		t.Fatalf("subscription = %q", got)
	}
}
