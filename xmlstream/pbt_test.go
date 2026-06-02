package xmlstream

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

func TestParser_Property_SelfClosingRoundtrip(t *testing.T) {
	nameGen := rapid.StringMatching(`[a-z][a-z0-9-]{0,20}`)
	attrNameGen := rapid.StringMatching(`[a-z]{1,8}`)
	attrValGen := rapid.StringMatching(`[a-zA-Z0-9]{1,16}`)

	rapid.Check(t, func(t *rapid.T) {
		name := nameGen.Draw(t, "name")
		attrName := attrNameGen.Draw(t, "attrName")
		attrVal := attrValGen.Draw(t, "attrVal")

		xml := fmt.Sprintf("<%s %s='%s'/>", name, attrName, attrVal)
		p := NewParser()
		p.Feed([]byte(xml))

		tok, err := p.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if tok.Kind != KindElement {
			t.Fatalf("kind = %v", tok.Kind)
		}
		if tok.Name != name {
			t.Fatalf("name = %q want %q", tok.Name, name)
		}
		if got := tok.AttrValue(attrName); got != attrVal {
			t.Fatalf("attr %q = %q want %q", attrName, got, attrVal)
		}
	})
}
