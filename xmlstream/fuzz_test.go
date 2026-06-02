package xmlstream

import "testing"

func FuzzParserFeed(f *testing.F) {
	f.Add([]byte("<?xml version='1.0'?><stream:stream xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>"))
	f.Add([]byte("<message><body>x</body></message>"))
	f.Add([]byte("<!--nope-->"))

	f.Fuzz(func(t *testing.T, data []byte) {
		p := NewParser()
		p.Feed(data)
		for i := 0; i < 32; i++ {
			_, err := p.Next()
			if err == ErrNeedMore {
				return
			}
			if err != nil {
				return
			}
		}
	})
}
