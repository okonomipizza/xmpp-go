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
