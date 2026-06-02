package protocol

import (
	"strings"
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func TestClientStreamOpenBytes_RFC6120(t *testing.T) {
	j, err := jid.Parse("juliet@im.example.com")
	if err != nil {
		t.Fatal(err)
	}
	data, err := ClientStreamOpenBytes(Config{JID: j, Lang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		"<?xml version='1.0'?>",
		"<stream:stream",
		"from='juliet@im.example.com'",
		"to='im.example.com'",
		"version='1.0'",
		"xml:lang='en'",
		"xmlns='jabber:client'",
		"xmlns:stream='http://etherx.jabber.org/streams'",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
}

func TestClientStreamOpenBytes_Errors(t *testing.T) {
	_, err := ClientStreamOpenBytes(Config{})
	if err == nil {
		t.Fatal("expected error for empty JID")
	}

	j, _ := jid.Parse("user@example.com/resource")
	_, err = ClientStreamOpenBytes(Config{JID: j})
	if err == nil {
		t.Fatal("expected error for full JID")
	}
}

func TestClientStreamOpenBytes_InvalidLang(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	_, err := ClientStreamOpenBytes(Config{JID: j, Lang: "en' x='y"})
	if err == nil {
		t.Fatal("expected error for invalid lang")
	}
}

func TestClientStreamOpenBytes_DefaultLang(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	data, err := ClientStreamOpenBytes(Config{JID: j})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "xml:lang='en'") {
		t.Fatal("expected default lang en")
	}
}
