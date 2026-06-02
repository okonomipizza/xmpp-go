package protocol

import (
	"testing"

	"github.com/okonomipizza/xmpp-go/jid"
)

func FuzzConnectionReceive(f *testing.F) {
	j, _ := jid.Parse("user@example.com")
	cfg := Config{JID: j}

	f.Add([]byte(serverOpen))
	f.Add([]byte("<message><body>x</body></message>"))

	f.Fuzz(func(t *testing.T, data []byte) {
		conn := NewConnection(cfg)
		_ = conn.Start()
		_, _ = conn.Receive(data)
	})
}

func FuzzRosterParsing(f *testing.F) {
	account, _ := jid.Parse("user@example.com")
	f.Add([]byte(`<query xmlns='jabber:iq:roster'><item jid='a@example.com'/></query>`))
	f.Add([]byte(`<query xmlns='jabber:iq:roster'><item jid='juliet@example.com' subscription='none' ask='subscribe'/></query>`))
	f.Add([]byte(`<wrapper><query xmlns='jabber:iq:roster'><item jid='a@example.com' subscription='bogus'/></query></wrapper>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseRosterItems(data)
		_, _ = ParseRosterPush(IQEvent{Type: "set", Payload: data}, account)
	})
}
