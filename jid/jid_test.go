package jid

import (
	"strings"
	"testing"
)

// TestParse_ValidJIDs tests valid JIDs based on RFC 7622 Table 1
func TestParse_ValidJIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		local    string
		domain   string
		resource string
	}{
		{
			name:     "bare JID",
			input:    "juliet@example.com",
			local:    "juliet",
			domain:   "example.com",
			resource: "",
		},
		{
			name:     "full JID",
			input:    "juliet@example.com/foo",
			local:    "juliet",
			domain:   "example.com",
			resource: "foo",
		},
		{
			name:     "resource with space",
			input:    "juliet@example.com/foo bar",
			local:    "juliet",
			domain:   "example.com",
			resource: "foo bar",
		},
		{
			name:     "resource with @",
			input:    "juliet@example.com/foo@bar",
			local:    "juliet",
			domain:   "example.com",
			resource: "foo@bar",
		},
		{
			name:     "domain only",
			input:    "example.com",
			local:    "",
			domain:   "example.com",
			resource: "",
		},
		{
			name:     "domain and resource",
			input:    "example.com/foobar",
			local:    "",
			domain:   "example.com",
			resource: "foobar",
		},
		{
			name:     "resource with @ in domain JID (RFC 7622 Example 15)",
			input:    "a.example.com/b@example.net",
			local:    "",
			domain:   "a.example.com",
			resource: "b@example.net",
		},
		{
			name:     "unicode local (Greek pi)",
			input:    "π@example.com",
			local:    "π",
			domain:   "example.com",
			resource: "",
		},
		{
			name:     "unicode resource (Chess king)",
			input:    "king@example.com/♚",
			local:    "king",
			domain:   "example.com",
			resource: "♚",
		},
		{
			name:     "IPv4 address domain",
			input:    "user@192.168.1.1",
			local:    "user",
			domain:   "192.168.1.1",
			resource: "",
		},
		{
			name:     "IPv6 address domain",
			input:    "user@[::1]",
			local:    "user",
			domain:   "[::1]",
			resource: "",
		},
		{
			name:     "resource with /",
			input:    "user@example.com/foo/bar/baz",
			local:    "user",
			domain:   "example.com",
			resource: "foo/bar/baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jid, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) should not return error: %v", tt.input, err)
			}
			if jid.Local() != tt.local {
				t.Errorf("Local() = %q, want %q", jid.Local(), tt.local)
			}
			if jid.Domain() != tt.domain {
				t.Errorf("Domain() = %q, want %q", jid.Domain(), tt.domain)
			}
			if jid.Resource() != tt.resource {
				t.Errorf("Resource() = %q, want %q", jid.Resource(), tt.resource)
			}
		})
	}
}

// TestParse_InvalidJIDs tests invalid JIDs based on RFC 7622 Table 2
func TestParse_InvalidJIDs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "empty JID",
			input:       "",
			expectedErr: ErrEmptyJID,
		},
		{
			name:        "local with quotation mark",
			input:       `"juliet"@example.com`,
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "local with ampersand",
			input:       "foo&bar@example.com",
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "local with apostrophe",
			input:       "foo'bar@example.com",
			expectedErr: ErrForbiddenChar,
		},
		// Note: "foo/bar@example.com" is parsed as domain=foo, resource=bar@example.com
		// due to RFC 7622 parsing order, so we cannot test / in localpart
		{
			name:        "local with colon",
			input:       "foo:bar@example.com",
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "local with less-than",
			input:       "foo<bar@example.com",
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "local with greater-than",
			input:       "foo>bar@example.com",
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "empty local (RFC 7622 Example 19)",
			input:       "@example.com",
			expectedErr: ErrEmptyLocal,
		},
		{
			name:        "empty resource",
			input:       "juliet@example.com/",
			expectedErr: ErrEmptyResource,
		},
		{
			name:        "empty local and empty resource",
			input:       "@example.com/",
			expectedErr: ErrEmptyLocal,
		},
		{
			name:        "no domain (RFC 7622 Example 22)",
			input:       "juliet@",
			expectedErr: ErrEmptyDomain,
		},
		{
			name:        "no domain, resource only (RFC 7622 Example 23)",
			input:       "/foobar",
			expectedErr: ErrEmptyDomain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Fatalf("Parse(%q) should return error", tt.input)
			}
			if err != tt.expectedErr {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, err, tt.expectedErr)
			}
		})
	}
}

// TestParse_LengthLimits tests length limits
func TestParse_LengthLimits(t *testing.T) {
	// 1023 octets is the maximum allowed
	maxLocal := strings.Repeat("a", 1023)
	maxDomain := strings.Repeat("a", 1023)
	maxResource := strings.Repeat("a", 1023)

	t.Run("max local", func(t *testing.T) {
		jid, err := Parse(maxLocal + "@example.com")
		if err != nil {
			t.Fatalf("max length local should be allowed: %v", err)
		}
		if len(jid.Local()) != 1023 {
			t.Errorf("Local() length = %d, want 1023", len(jid.Local()))
		}
	})

	t.Run("max domain", func(t *testing.T) {
		jid, err := Parse(maxDomain)
		if err != nil {
			t.Fatalf("max length domain should be allowed: %v", err)
		}
		if len(jid.Domain()) != 1023 {
			t.Errorf("Domain() length = %d, want 1023", len(jid.Domain()))
		}
	})

	t.Run("max resource", func(t *testing.T) {
		jid, err := Parse("example.com/" + maxResource)
		if err != nil {
			t.Fatalf("max length resource should be allowed: %v", err)
		}
		if len(jid.Resource()) != 1023 {
			t.Errorf("Resource() length = %d, want 1023", len(jid.Resource()))
		}
	})

	// 1024 octets exceeds the limit
	tooLongLocal := strings.Repeat("a", 1024)
	tooLongDomain := strings.Repeat("a", 1024)
	tooLongResource := strings.Repeat("a", 1024)

	t.Run("local too long", func(t *testing.T) {
		_, err := Parse(tooLongLocal + "@example.com")
		if err != ErrLocalTooLong {
			t.Errorf("1024 octet local should return ErrLocalTooLong, got %v", err)
		}
	})

	t.Run("domain too long", func(t *testing.T) {
		_, err := Parse(tooLongDomain)
		if err != ErrDomainTooLong {
			t.Errorf("1024 octet domain should return ErrDomainTooLong, got %v", err)
		}
	})

	t.Run("resource too long", func(t *testing.T) {
		_, err := Parse("example.com/" + tooLongResource)
		if err != ErrResourceTooLong {
			t.Errorf("1024 octet resource should return ErrResourceTooLong, got %v", err)
		}
	})
}

// TestJID_String tests String() method
func TestJID_String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare JID",
			input:    "user@example.com",
			expected: "user@example.com",
		},
		{
			name:     "full JID",
			input:    "user@example.com/resource",
			expected: "user@example.com/resource",
		},
		{
			name:     "domain only",
			input:    "example.com",
			expected: "example.com",
		},
		{
			name:     "domain and resource",
			input:    "example.com/resource",
			expected: "example.com/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jid, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.input, err)
			}
			if jid.String() != tt.expected {
				t.Errorf("String() = %q, want %q", jid.String(), tt.expected)
			}
		})
	}
}

// TestJID_Bare tests Bare() method
func TestJID_Bare(t *testing.T) {
	jid, _ := Parse("user@example.com/resource")
	bare := jid.Bare()

	if bare.Local() != "user" {
		t.Errorf("Bare().Local() = %q, want %q", bare.Local(), "user")
	}
	if bare.Domain() != "example.com" {
		t.Errorf("Bare().Domain() = %q, want %q", bare.Domain(), "example.com")
	}
	if bare.Resource() != "" {
		t.Errorf("Bare().Resource() = %q, want empty", bare.Resource())
	}
	if bare.String() != "user@example.com" {
		t.Errorf("Bare().String() = %q, want %q", bare.String(), "user@example.com")
	}
}

// TestJID_Equal tests Equal() method
func TestJID_Equal(t *testing.T) {
	jid1, _ := Parse("user@example.com/resource")
	jid2, _ := Parse("user@example.com/resource")
	jid3, _ := Parse("user@example.com/other")
	jid4, _ := Parse("other@example.com/resource")

	if !jid1.Equal(jid2) {
		t.Errorf("identical JIDs should be equal")
	}
	if jid1.Equal(jid3) {
		t.Errorf("JIDs with different resource should not be equal")
	}
	if jid1.Equal(jid4) {
		t.Errorf("JIDs with different local should not be equal")
	}
}

// TestJID_IsBareAndIsFull tests IsBare() and IsFull() methods
func TestJID_IsBareAndIsFull(t *testing.T) {
	bareJID, _ := Parse("user@example.com")
	fullJID, _ := Parse("user@example.com/resource")

	if !bareJID.IsBare() {
		t.Errorf("JID without resource should be bare")
	}
	if bareJID.IsFull() {
		t.Errorf("JID without resource should not be full")
	}
	if fullJID.IsBare() {
		t.Errorf("JID with resource should not be bare")
	}
	if !fullJID.IsFull() {
		t.Errorf("JID with resource should be full")
	}
}
