package protocol

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

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

func TestParseRosterItems_FuzzRegression_9af5fe10a8e08f60(t *testing.T) {
	data, err := os.ReadFile("testdata/fuzz/FuzzRosterParsing/9af5fe10a8e08f60")
	if err != nil {
		t.Fatal(err)
	}
	const want = `<wrapper><query xmlns='jabber:iq:roster'><='a@example.com' ''/></query</wrapper>`
	if !strings.Contains(string(data), want) {
		t.Fatalf("corpus changed: %s", data)
	}
	fuzzRegressionNoHang(t, []byte(want))
}

func TestParseRosterItems_FuzzRegression_af57aafd0ebe4fef(t *testing.T) {
	data, err := os.ReadFile("testdata/fuzz/FuzzRosterParsing/af57aafd0ebe4fef")
	if err != nil {
		t.Fatal(err)
	}
	const want = `<query ''><''></query>`
	if !strings.Contains(string(data), want) {
		t.Fatalf("corpus changed: %s", data)
	}
	fuzzRegressionNoHang(t, []byte(want))
}

func fuzzRegressionNoHang(t *testing.T, input []byte) {
	t.Helper()
	account := mustJID(t, "user@example.com")
	done := make(chan struct{})
	go func() {
		_, _ = ParseRosterItems(input)
		_, _ = ParseRosterPush(IQEvent{Type: "set", Payload: input}, account)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("roster parsing hung on fuzz regression input")
	}
}

func mustJID(t *testing.T, s string) jid.JID {
	t.Helper()
	j, err := jid.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return j
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

func TestParseRosterItems_InvalidSubscription(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com' subscription='bogus'/></query>`)
	items, err := ParseRosterItems(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Subscription != "" {
		t.Fatalf("items = %+v", items)
	}
}

func TestParseRosterItems_RemoveIgnored(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com' subscription='remove'/></query>`)
	items, err := ParseRosterItems(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Subscription != "" {
		t.Fatalf("items = %+v", items)
	}
}

func TestParseRosterPush_InvalidSubscription(t *testing.T) {
	ev := IQEvent{
		Type:    "set",
		Payload: []byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com' subscription='bogus'/></query>`),
	}
	account, _ := jid.Parse("u@example.com")
	got, err := ParseRosterPush(ev, account)
	if err != nil {
		t.Fatal(err)
	}
	if got.Subscription != "" {
		t.Fatalf("subscription = %q", got.Subscription)
	}
}

func TestParseRosterPush_QueryAsChild(t *testing.T) {
	ev := IQEvent{
		Type: "set",
		Payload: []byte(`<wrapper><query xmlns='jabber:iq:roster'>` +
			`<item jid='a@example.com' subscription='both'/></query></wrapper>`),
	}
	account, _ := jid.Parse("u@example.com")
	got, err := ParseRosterPush(ev, account)
	if err != nil {
		t.Fatal(err)
	}
	if got.Subscription != "both" {
		t.Fatalf("subscription = %q", got.Subscription)
	}
}

func TestParseRosterPush_NoQuery(t *testing.T) {
	ev := IQEvent{Type: "set", Payload: []byte(`<ping xmlns='urn:xmpp:ping'/>`)}
	account, _ := jid.Parse("u@example.com")
	_, err := ParseRosterPush(ev, account)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterPush_RemoveSubscription(t *testing.T) {
	ev := IQEvent{
		Type:    "set",
		Payload: []byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com' subscription='remove'/></query>`),
	}
	account, _ := jid.Parse("u@example.com")
	got, err := ParseRosterPush(ev, account)
	if err != nil {
		t.Fatal(err)
	}
	if got.Subscription != "remove" {
		t.Fatalf("subscription = %q", got.Subscription)
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

func TestRosterSetItemBytes_RFC6121Section2_3_1(t *testing.T) {
	j, _ := jid.Parse("nurse@example.com")
	item, err := RosterSetItemBytes(RosterSetItem{
		JID:    j,
		Name:   "Nurse",
		Groups: []string{"Servants"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(item)
	for _, want := range []string{
		"jid='nurse@example.com'",
		"name='Nurse'",
		"<group>Servants</group>",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
	if strings.Contains(s, "subscription=") {
		t.Fatalf("unexpected subscription in %s", s)
	}
}

func TestRosterSetItemBytes_RFC6121Section2_5_1(t *testing.T) {
	j, _ := jid.Parse("nurse@example.com")
	item, err := RosterSetItemBytes(RosterSetItem{JID: j, Subscription: "remove"})
	if err != nil {
		t.Fatal(err)
	}
	want := "<item jid='nurse@example.com' subscription='remove'/>"
	if string(item) != want {
		t.Fatalf("got %s", item)
	}
}

func TestRosterIQSetBytes_RFC6121Section2_3_1(t *testing.T) {
	from, _ := jid.Parse("juliet@example.com/balcony")
	j, _ := jid.Parse("nurse@example.com")
	data, err := RosterIQSetBytes(from, "ph1xaz53", RosterSetItem{JID: j, Name: "Nurse", Groups: []string{"Servants"}})
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		"from='juliet@example.com/balcony'",
		"id='ph1xaz53'",
		"type='set'",
		"<query xmlns='jabber:iq:roster'>",
		"jid='nurse@example.com'",
		"<group>Servants</group>",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestRosterSetItemBytes_RemoveWithName(t *testing.T) {
	j, _ := jid.Parse("a@example.com")
	_, err := RosterSetItemBytes(RosterSetItem{JID: j, Subscription: "remove", Name: "A"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRosterSetItemBytes_FullJID(t *testing.T) {
	j, _ := jid.Parse("a@example.com/res")
	_, err := RosterSetItemBytes(RosterSetItem{JID: j})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRosterSetItemBytes_EmptyJID(t *testing.T) {
	_, err := RosterSetItemBytes(RosterSetItem{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRosterSetItemBytes_InvalidSubscription(t *testing.T) {
	j, _ := jid.Parse("a@example.com")
	for _, sub := range []string{"invalid", "both", "none", "to", "from"} {
		_, err := RosterSetItemBytes(RosterSetItem{JID: j, Subscription: sub})
		if err == nil {
			t.Fatalf("subscription %q: expected error", sub)
		}
	}
}

func TestRosterSetItemBytes_EmptyGroup(t *testing.T) {
	j, _ := jid.Parse("a@example.com")
	_, err := RosterSetItemBytes(RosterSetItem{JID: j, Groups: []string{""}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRosterSetItemBytes_DuplicateGroup(t *testing.T) {
	j, _ := jid.Parse("a@example.com")
	_, err := RosterSetItemBytes(RosterSetItem{JID: j, Groups: []string{"Friends", "Friends"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRosterQuerySetBytes_EmptyJID(t *testing.T) {
	_, err := RosterQuerySetBytes(RosterSetItem{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterPush_RFC6121Section2_1_6(t *testing.T) {
	const pushIQ = `<iq id='b89c5r7ib574' to='romeo@example.net/foo' type='set'>` +
		`<query xmlns='jabber:iq:roster'>` +
		`<item ask='subscribe' jid='juliet@example.com' subscription='none'/>` +
		`</query></iq>`
	tok := iqToken(t, pushIQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	account, _ := jid.Parse("romeo@example.net")
	got, err := ParseRosterPush(ev, account)
	if err != nil {
		t.Fatal(err)
	}
	j, _ := jid.Parse("juliet@example.com")
	if got.JID != j || got.Subscription != "none" || got.Ask != "subscribe" {
		t.Fatalf("item = %+v", got)
	}
}

func TestParseRosterPush_FromBareJID(t *testing.T) {
	const pushIQ = `<iq from='romeo@example.net' id='p1' to='romeo@example.net/foo' type='set'>` +
		`<query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query></iq>`
	tok := iqToken(t, pushIQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	account, _ := jid.Parse("romeo@example.net")
	if _, err := ParseRosterPush(ev, account); err != nil {
		t.Fatal(err)
	}
}

func TestParseRosterPush_WrongFrom(t *testing.T) {
	const pushIQ = `<iq from='other@example.net' id='p1' to='romeo@example.net/foo' type='set'>` +
		`<query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query></iq>`
	tok := iqToken(t, pushIQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	account, _ := jid.Parse("romeo@example.net")
	_, err = ParseRosterPush(ev, account)
	if !errors.Is(err, ErrRosterPushIgnored) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRosterPush_FromFullJID(t *testing.T) {
	const pushIQ = `<iq from='romeo@example.net/attacker' id='p1' to='romeo@example.net/foo' type='set'>` +
		`<query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query></iq>`
	tok := iqToken(t, pushIQ)
	ev, err := ParseInboundIQ(tok)
	if err != nil {
		t.Fatal(err)
	}
	account, _ := jid.Parse("romeo@example.net")
	_, err = ParseRosterPush(ev, account)
	if !errors.Is(err, ErrRosterPushIgnored) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRosterPush_MultipleItems(t *testing.T) {
	payload := []byte(`<query xmlns='jabber:iq:roster'>` +
		`<item jid='a@example.com'/><item jid='b@example.com'/>` +
		`</query>`)
	ev := IQEvent{Type: "set", Payload: payload}
	account, _ := jid.Parse("u@example.com")
	_, err := ParseRosterPush(ev, account)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterPush_NotSet(t *testing.T) {
	ev := IQEvent{Type: "result", Payload: []byte(`<query xmlns='jabber:iq:roster'/>`)}
	account, _ := jid.Parse("u@example.com")
	_, err := ParseRosterPush(ev, account)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRosterPush_InvalidAccountBare(t *testing.T) {
	ev := IQEvent{
		Type:    "set",
		Payload: []byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query>`),
	}
	full, _ := jid.Parse("u@example.com/r")
	_, err := ParseRosterPush(ev, full)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendRosterSet(t *testing.T) {
	j, _ := jid.Parse("juliet@example.com/balcony")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	contact, _ := jid.Parse("nurse@example.com")
	if err := conn.SendRosterSet("ph1xaz53", RosterSetItem{JID: contact, Name: "Nurse"}); err != nil {
		t.Fatal(err)
	}
	data := conn.BytesToSend()
	if data == nil || !strings.Contains(string(data), "type='set'") {
		t.Fatalf("data: %s", data)
	}
}

func TestConnection_SendRosterRemove(t *testing.T) {
	j, _ := jid.Parse("juliet@example.com/balcony")
	conn := NewConnection(Config{JID: j})
	conn.state = StateReady
	conn.boundJID = j
	contact, _ := jid.Parse("nurse@example.com")
	if err := conn.SendRosterRemove("hm4hs97y", contact); err != nil {
		t.Fatal(err)
	}
	data := conn.BytesToSend()
	if data == nil || !strings.Contains(string(data), "subscription='remove'") {
		t.Fatalf("data: %s", data)
	}
}

func TestConnection_SendRosterSetNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.boundJID = j
	contact, _ := jid.Parse("a@example.com")
	if err := conn.SendRosterSet("s1", RosterSetItem{JID: contact}); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnection_SendRosterRemoveNotReady(t *testing.T) {
	j, _ := jid.Parse("u@example.com/r")
	conn := NewConnection(Config{JID: j})
	conn.boundJID = j
	contact, _ := jid.Parse("a@example.com")
	if err := conn.SendRosterRemove("r1", contact); err == nil {
		t.Fatal("expected error")
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
