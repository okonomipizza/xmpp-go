package protocol

import (
	"testing"
	"time"

	"github.com/okonomipizza/xmpp-go/xmlstream"
)

func TestParseStreamFeatures_STARTTLSRequired(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw:  []byte(featuresTLS),
	}
	f := ParseStreamFeatures(tok)
	if !f.StartTLSOffered || !f.StartTLSRequired {
		t.Fatalf("features = %+v", f)
	}
}

func TestParseStreamFeatures_Mechanisms(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw: []byte(`<stream:features><mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>` +
			`<mechanism>PLAIN</mechanism><mechanism>SCRAM-SHA-1</mechanism></mechanisms></stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if len(f.Mechanisms) != 2 || f.Mechanisms[0] != "PLAIN" {
		t.Fatalf("mechanisms = %v", f.Mechanisms)
	}
}

func TestParseStreamFeatures_NoFalsePositiveStartTLS(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw: []byte(`<stream:features><note>` +
			`documentation mentions &lt;starttls and &lt;required` +
			`</note></stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if f.StartTLSOffered || f.StartTLSRequired {
		t.Fatalf("features = %+v", f)
	}
}

func TestParseStreamFeatures_NoTLS(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw:  []byte(`<stream:features><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/></stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if f.StartTLSOffered {
		t.Fatal("unexpected STARTTLS")
	}
	if !f.BindOffered {
		t.Fatal("expected bind")
	}
}

func TestParseDirectChildElements_nested(t *testing.T) {
	raw := []byte(`<mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>` +
		`<mechanism>PLAIN</mechanism><mechanism>X</mechanism></mechanisms>`)
	children := parseDirectChildElements(raw)
	if len(children) != 2 || children[0].name != "mechanism" {
		t.Fatalf("children = %+v", children)
	}
}

func TestParseStreamFeatures_Malformed(t *testing.T) {
	tok := xmlstream.Token{Kind: xmlstream.KindElement, Name: "stream:features", Raw: []byte("<stream:features>")}
	f := ParseStreamFeatures(tok)
	if f.StartTLSOffered || f.BindOffered || len(f.Mechanisms) > 0 {
		t.Fatalf("features = %+v", f)
	}
}

func TestParseStreamFeatures_WithPI(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw: []byte(`<stream:features><?pi target="x"?>` +
			`<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/></stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if !f.StartTLSOffered {
		t.Fatalf("features = %+v", f)
	}
}

func TestParseStreamFeatures_WithComment(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw: []byte(`<stream:features><!-- offer bind -->` +
			`<bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/></stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if !f.BindOffered {
		t.Fatalf("features = %+v", f)
	}
}

func TestSkipNonElementMarkup_Incomplete(t *testing.T) {
	if end := skipNonElementMarkup([]byte(`<?pi`), 0); end != 4 {
		t.Fatalf("end = %d", end)
	}
}

func TestFindOpenTagEnd_QuotedGt(t *testing.T) {
	tag := []byte(`<elem attr=">">`)
	end, selfClosing, ok := findOpenTagEnd(tag, 0)
	if !ok || selfClosing || end != len(tag)-1 {
		t.Fatalf("end=%d selfClosing=%v ok=%v", end, selfClosing, ok)
	}
}

func TestParseDirectChildElements_MismatchedClose(t *testing.T) {
	raw := []byte(`<outer><inner></wrong></outer>`)
	children := parseDirectChildElements(raw)
	if len(children) != 0 {
		t.Fatalf("children = %+v", children)
	}
}

func TestParseDirectChildElements_IncompleteOpen(t *testing.T) {
	if children := parseDirectChildElements([]byte(`<outer><inner`)); len(children) != 0 {
		t.Fatalf("children = %+v", children)
	}
}

func TestParseDirectChildElements_UnclosedChildTerminates(t *testing.T) {
	// FuzzRosterParsing corpus af57aafd0ebe4fef: must not hang on malformed child close tags
	raw := []byte(`<query ''><''></query>`)
	done := make(chan struct{})
	go func() {
		_ = parseDirectChildElements(raw)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("parseDirectChildElements hung")
	}
}

func TestParseStreamFeatures_AttributeWithGt(t *testing.T) {
	tok := xmlstream.Token{
		Kind: xmlstream.KindElement,
		Name: "stream:features",
		Raw: []byte(`<stream:features>` +
			`<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls' hint=">"/>` +
			`</stream:features>`),
	}
	f := ParseStreamFeatures(tok)
	if !f.StartTLSOffered {
		t.Fatalf("features = %+v", f)
	}
}
