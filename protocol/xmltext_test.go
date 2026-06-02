package protocol

import "testing"

func TestUnescapeXMLText(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"Tom &amp; Jerry", "Tom & Jerry"},
		{"a &lt; b", "a < b"},
		{"a &gt; b", "a > b"},
		{`&apos;x&apos;`, "'x'"},
		{`&quot;hi&quot;`, `"hi"`},
		{"&amp;amp;", "&amp;"}, // 二重エスケープは 1 段階のみ復元
		{"unknown &entity; here", "unknown &entity; here"},
		{"trailing &amp", "trailing &amp"},
	}
	for _, tc := range tests {
		if got := unescapeXMLText(tc.in); got != tc.want {
			t.Fatalf("unescapeXMLText(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEscapeUnescapeXMLText_Roundtrip(t *testing.T) {
	const in = "Tom & Jerry <3 >0"
	got := unescapeXMLText(escapeXMLText(in))
	if got != in {
		t.Fatalf("roundtrip = %q", got)
	}
}
