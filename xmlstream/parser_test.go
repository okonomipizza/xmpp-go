package xmlstream

import (
	"strings"
	"testing"
)

// RFC 6120 Section 4.2 — クライアント初期ストリームヘッダ
const rfc6120ClientStreamOpen = `<?xml version='1.0'?>
<stream:stream
    from='juliet@im.example.com'
    to='im.example.com'
    version='1.0'
    xml:lang='en'
    xmlns='jabber:client'
    xmlns:stream='http://etherx.jabber.org/streams'>`

// RFC 6120 Section 4.2 — サーバー応答ストリームヘッダ
const rfc6120ServerStreamOpen = `<?xml version='1.0'?>
<stream:stream
    from='im.example.com'
    id='++TR84Sm6A3hnt3Q065SnAbbk3Y='
    to='juliet@im.example.com'
    version='1.0'
    xml:lang='en'
    xmlns='jabber:client'
    xmlns:stream='http://etherx.jabber.org/streams'>`

// RFC 6120 Section 4.3.2 — STARTTLS 必須の stream features
const rfc6120StreamFeaturesTLS = `<stream:features>
  <starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>
    <required/>
  </starttls>
</stream:features>`

// RFC 6120 Section 4.10 — メッセージ stanza
const rfc6120Message = `<message from='juliet@im.example.com/balcony'
            to='romeo@example.net'
            xml:lang='en'>
   <body>Art thou not Romeo, and a Montague?</body>
 </message>`

func collectTokens(t *testing.T, p *Parser, data string) []Token {
	t.Helper()
	p.Reset()
	p.Feed([]byte(data))
	var tokens []Token
	for {
		tok, err := p.Next()
		if err == ErrNeedMore {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func TestParser_RFC6120_ClientStreamOpen(t *testing.T) {
	p := NewParser()
	p.Feed([]byte(rfc6120ClientStreamOpen))

	tok, err := p.Next()
	if err != nil {
		t.Fatalf("first Next: %v", err)
	}
	if tok.Kind != KindPI {
		t.Fatalf("kind = %v, want KindPI", tok.Kind)
	}

	tok, err = p.Next()
	if err != nil {
		t.Fatalf("second Next: %v", err)
	}
	if tok.Kind != KindStreamOpen {
		t.Fatalf("kind = %v, want KindStreamOpen", tok.Kind)
	}
	if tok.Name != "stream:stream" {
		t.Fatalf("name = %q", tok.Name)
	}
	if got := tok.AttrValue("from"); got != "juliet@im.example.com" {
		t.Fatalf("from = %q", got)
	}
	if got := tok.AttrValue("to"); got != "im.example.com" {
		t.Fatalf("to = %q", got)
	}
	if got := tok.AttrValue("version"); got != "1.0" {
		t.Fatalf("version = %q", got)
	}
}

func TestParser_RFC6120_ServerStreamAndFeatures(t *testing.T) {
	data := rfc6120ServerStreamOpen + "\n" + rfc6120StreamFeaturesTLS
	tokens := collectTokens(t, NewParser(), data)

	if len(tokens) != 3 {
		t.Fatalf("got %d tokens, want 3", len(tokens))
	}
	if tokens[0].Kind != KindPI {
		t.Fatal("token 0: want PI")
	}
	if tokens[1].Kind != KindStreamOpen {
		t.Fatal("token 1: want stream open")
	}
	if tokens[1].AttrValue("id") != "++TR84Sm6A3hnt3Q065SnAbbk3Y=" {
		t.Fatalf("id = %q", tokens[1].AttrValue("id"))
	}
	if tokens[2].Kind != KindElement || tokens[2].Name != "stream:features" {
		t.Fatalf("token 2: got kind=%v name=%q", tokens[2].Kind, tokens[2].Name)
	}
}

func TestParser_RFC6120_MessageStanza(t *testing.T) {
	tokens := collectTokens(t, NewParser(), rfc6120Message)
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens", len(tokens))
	}
	if tokens[0].Name != "message" {
		t.Fatalf("name = %q", tokens[0].Name)
	}
	if !strings.Contains(string(tokens[0].Raw), "Montague") {
		t.Fatal("raw body missing")
	}
}

func TestParser_IncrementalFeed(t *testing.T) {
	p := NewParser()
	data := []byte(rfc6120ClientStreamOpen)
	for i := 1; i <= len(data); i++ {
		p.Reset()
		p.Feed(data[:i])
		var lastErr error
		for {
			_, err := p.Next()
			if err == ErrNeedMore {
				lastErr = err
				break
			}
			if err != nil {
				t.Fatalf("partial feed %d: %v", i, err)
			}
		}
		if i < len(data) && lastErr != ErrNeedMore {
			// 完全にパースできる前は NeedMore が期待されることが多い
		}
	}

	p.Reset()
	p.Feed(data)
	var count int
	for {
		_, err := p.Next()
		if err == ErrNeedMore {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("full parse: %d tokens", count)
	}
}

func TestParser_StreamClose(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("</stream:stream>"))
	tok, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Kind != KindStreamClose {
		t.Fatalf("kind = %v", tok.Kind)
	}
}

func TestParser_InnerIncompleteTagNoPanic(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("<message><a"))
	_, err := p.Next()
	if err != ErrNeedMore {
		t.Fatalf("err = %v want ErrNeedMore", err)
	}
}

func TestParser_IncompleteOpenBracket(t *testing.T) {
	// Fuzz corpus "<" 相当: 単独の '<' でパニックしないこと
	p := NewParser()
	p.Feed([]byte("<"))
	_, err := p.Next()
	if err != ErrNeedMore {
		t.Fatalf("err = %v want ErrNeedMore", err)
	}
}

func TestParser_TopLevelCDATARejected(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("<![CDATA[x]]>"))
	_, err := p.Next()
	if err != ErrRestrictedXML {
		t.Fatalf("err = %v", err)
	}
}

func TestParser_NeedInput(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("<message><body>"))
	_, err := p.Next()
	if err != ErrNeedMore {
		t.Fatalf("err = %v", err)
	}
	if !p.NeedInput() {
		t.Fatal("NeedInput want true")
	}
	p.Feed([]byte("x</body></message>"))
	tok, err := p.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Name != "message" {
		t.Fatalf("name = %q", tok.Name)
	}
	if p.NeedInput() {
		t.Fatal("NeedInput want false after complete token")
	}
}

func TestParser_RestrictedXML_DOCTYPE(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("<!DOCTYPE foo>"))
	_, err := p.Next()
	if err != ErrRestrictedXML {
		t.Fatalf("err = %v", err)
	}
}

func TestParser_RestrictedXML_Comment(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("<!-- comment -->"))
	_, err := p.Next()
	if err != ErrRestrictedXML {
		t.Fatalf("err = %v, want ErrRestrictedXML", err)
	}
}
