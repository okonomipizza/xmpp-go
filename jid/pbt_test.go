package jid

import (
	"testing"

	"pgregory.net/rapid"
)

// TestParse_Property_NoPanic verifies Parse does not panic on arbitrary input
func TestParse_Property_NoPanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := rapid.String().Draw(t, "input")
		// Parse should return error or succeed, never panic
		_, _ = Parse(input)
	})
}

// TestParse_Property_Roundtrip verifies parsed JID can be converted back to string
func TestParse_Property_Roundtrip(t *testing.T) {
	// Generator for valid local part (no forbidden chars)
	validLocalGen := rapid.StringMatching(`[a-zA-Z0-9._-]{1,100}`)
	// Generator for valid domain part
	validDomainGen := rapid.StringMatching(`[a-zA-Z0-9.-]{1,100}\.[a-zA-Z]{2,10}`)
	// Generator for valid resource part (non-empty)
	validResourceGen := rapid.StringMatching(`[a-zA-Z0-9 ._@-]{1,100}`)

	rapid.Check(t, func(t *rapid.T) {
		local := validLocalGen.Draw(t, "local")
		domain := validDomainGen.Draw(t, "domain")

		// Test bare JID
		bareInput := local + "@" + domain
		bareJID, err := Parse(bareInput)
		if err != nil {
			return // skip if generated value is invalid
		}

		// Verify roundtrip
		if bareJID.String() != bareInput {
			t.Fatalf("bare JID roundtrip failed: input=%q, output=%q", bareInput, bareJID.String())
		}

		// Test full JID
		resource := validResourceGen.Draw(t, "resource")
		fullInput := local + "@" + domain + "/" + resource
		fullJID, err := Parse(fullInput)
		if err != nil {
			return
		}

		if fullJID.String() != fullInput {
			t.Fatalf("full JID roundtrip failed: input=%q, output=%q", fullInput, fullJID.String())
		}
	})
}

// TestJID_Property_BareRemovesOnlyResource verifies Bare() removes only resource
func TestJID_Property_BareRemovesOnlyResource(t *testing.T) {
	validLocalGen := rapid.StringMatching(`[a-zA-Z0-9._-]{1,50}`)
	validDomainGen := rapid.StringMatching(`[a-zA-Z0-9.-]{1,50}\.[a-zA-Z]{2,5}`)
	validResourceGen := rapid.StringMatching(`[a-zA-Z0-9._-]{1,50}`)

	rapid.Check(t, func(t *rapid.T) {
		local := validLocalGen.Draw(t, "local")
		domain := validDomainGen.Draw(t, "domain")
		resource := validResourceGen.Draw(t, "resource")

		fullInput := local + "@" + domain + "/" + resource
		fullJID, err := Parse(fullInput)
		if err != nil {
			return
		}

		bareJID := fullJID.Bare()

		// Local and domain should remain unchanged
		if bareJID.Local() != fullJID.Local() {
			t.Fatalf("Bare() changed local: %q -> %q", fullJID.Local(), bareJID.Local())
		}
		if bareJID.Domain() != fullJID.Domain() {
			t.Fatalf("Bare() changed domain: %q -> %q", fullJID.Domain(), bareJID.Domain())
		}

		// Resource should be empty
		if bareJID.Resource() != "" {
			t.Fatalf("Bare() did not remove resource: %q", bareJID.Resource())
		}
	})
}

// TestJID_Property_EqualSymmetry verifies Equal() is reflexive and symmetric
func TestJID_Property_EqualSymmetry(t *testing.T) {
	validLocalGen := rapid.StringMatching(`[a-zA-Z0-9._-]{1,50}`)
	validDomainGen := rapid.StringMatching(`[a-zA-Z0-9.-]{1,50}\.[a-zA-Z]{2,5}`)
	validResourceGen := rapid.StringMatching(`[a-zA-Z0-9._-]{0,50}`)

	rapid.Check(t, func(t *rapid.T) {
		local := validLocalGen.Draw(t, "local")
		domain := validDomainGen.Draw(t, "domain")
		resource := validResourceGen.Draw(t, "resource")

		var input string
		if resource == "" {
			input = local + "@" + domain
		} else {
			input = local + "@" + domain + "/" + resource
		}

		j1, err := Parse(input)
		if err != nil {
			return
		}

		j2, err := Parse(input)
		if err != nil {
			return
		}

		// Reflexive: j1 == j1
		if !j1.Equal(j1) {
			t.Fatalf("Equal() is not reflexive: %v", j1)
		}

		// Symmetric: j1 == j2 implies j2 == j1
		if j1.Equal(j2) != j2.Equal(j1) {
			t.Fatalf("Equal() is not symmetric: j1=%v, j2=%v", j1, j2)
		}
	})
}
