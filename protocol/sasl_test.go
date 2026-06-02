package protocol

import (
	"encoding/base64"
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func TestPLAINInitialResponse_Authz(t *testing.T) {
	got := PLAINInitialResponse("admin@example.com", "user", "pass")
	want := "\x00user\x00pass"
	if string(got) != "admin@example.com"+want {
		t.Fatalf("got %q", got)
	}
}

func TestPLAINInitialResponse_RFC6120(t *testing.T) {
	// RFC 6120 Section 6.4 の例: juliet / r0m30myr0m30
	got := PLAINInitialResponse("", "juliet", "r0m30myr0m30")
	want, _ := base64.StdEncoding.DecodeString("AGp1bGlldAByMG0zMG15cjBtMzA=")
	if string(got) != string(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestAuthElementBytes_PLAIN(t *testing.T) {
	initial := PLAINInitialResponse("", "juliet", "secret")
	data, err := AuthElementBytes("PLAIN", initial)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !containsAll(s, "auth", "PLAIN", saslNS) {
		t.Fatalf("auth element: %s", s)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if p == "" || !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexSub(s, sub) >= 0)
}

func indexSub(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestAuthenticate_PLAINRequiresPassword(t *testing.T) {
	j, _ := jid.Parse("user@example.com")
	conn := NewConnection(Config{JID: j})
	conn.state = StateNegotiating
	conn.features = StreamFeatures{Mechanisms: []string{"PLAIN"}}

	if err := conn.Authenticate("PLAIN"); err == nil {
		t.Fatal("expected error without password")
	}
}

func TestAuthElementBytes_InvalidMechanism(t *testing.T) {
	_, err := AuthElementBytes("bad mech", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSASLInitial_EmptyLocalpart(t *testing.T) {
	j, _ := jid.Parse("example.com")
	cfg := Config{JID: j, Password: "x"}
	_, err := saslInitial(cfg, "PLAIN")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectMechanism_NoneOffered(t *testing.T) {
	j, _ := jid.Parse("u@example.com")
	conn := NewConnection(Config{JID: j})
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1"}}
	if _, err := conn.selectMechanism(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateSASLMechanism_Invalid(t *testing.T) {
	if err := validateSASLMechanism("bad mech"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectMechanism_DefaultPLAIN(t *testing.T) {
	j, _ := jid.Parse("u@example.com")
	conn := NewConnection(Config{JID: j, Password: "p"})
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1", "PLAIN"}}
	m, err := conn.selectMechanism()
	if err != nil || m != "PLAIN" {
		t.Fatalf("mechanism = %q err = %v", m, err)
	}
}

func TestSelectMechanism(t *testing.T) {
	j, _ := jid.Parse("u@example.com")
	conn := NewConnection(Config{
		JID:            j,
		SASLMechanisms: []string{"PLAIN", "SCRAM-SHA-1"},
	})
	conn.features = StreamFeatures{Mechanisms: []string{"SCRAM-SHA-1", "PLAIN"}}
	m, err := conn.selectMechanism()
	if err != nil || m != "PLAIN" {
		t.Fatalf("mechanism = %q err = %v", m, err)
	}
}
