package protocol

import "testing"

func TestStartTLSCommandBytes_RFC6120(t *testing.T) {
	got := string(StartTLSCommandBytes())
	want := "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
