package jid

import "testing"

// FuzzParse verifies Parse does not panic on arbitrary input
func FuzzParse(f *testing.F) {
	// Seed corpus: valid JIDs
	f.Add("user@example.com")
	f.Add("user@example.com/resource")
	f.Add("example.com")
	f.Add("example.com/resource")
	f.Add("juliet@example.com/foo bar")
	f.Add("a.example.com/b@example.net")

	// Seed corpus: invalid JIDs
	f.Add("")
	f.Add("@")
	f.Add("/")
	f.Add("@example.com")
	f.Add("juliet@")
	f.Add("/foobar")
	f.Add("juliet@example.com/")

	// Seed corpus: boundary values
	f.Add("a")
	f.Add("@a")
	f.Add("a@")
	f.Add("a/")
	f.Add("/a")
	f.Add("a@b")
	f.Add("a@b/c")

	// Seed corpus: forbidden chars
	f.Add(`"user"@example.com`)
	f.Add("user&name@example.com")
	f.Add("user'name@example.com")
	f.Add("user:name@example.com")
	f.Add("user<name@example.com")
	f.Add("user>name@example.com")

	// Seed corpus: unicode
	f.Add("日本語@example.com")
	f.Add("user@example.com/日本語")
	f.Add("π@example.com")

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should return error or succeed, never panic
		jid, err := Parse(input)
		if err != nil {
			return
		}

		// If parse succeeds, accessing fields should not panic
		_ = jid.Local()
		_ = jid.Domain()
		_ = jid.Resource()
		_ = jid.String()
		_ = jid.Bare()
		_ = jid.IsBare()
		_ = jid.IsFull()
		_ = jid.IsEmpty()
		_ = jid.Equal(jid)
	})
}
