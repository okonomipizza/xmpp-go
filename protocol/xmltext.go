package protocol

import (
	"bytes"
	"strings"
)

// elementTextContent は要素の生 XML から最初のテキスト子ノードを返す。
func elementTextContent(raw []byte) string {
	start := bytes.IndexByte(raw, '>')
	if start < 0 {
		return ""
	}
	start++
	end := bytes.LastIndex(raw, []byte("</"))
	if end < 0 || end <= start {
		return ""
	}
	return string(bytes.TrimSpace(raw[start:end]))
}

// escapeXMLText は XML テキストノード用に &, <, > をエスケープする。
func escapeXMLText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// unescapeXMLText は XML テキストノード内の定義済みエンティティを復元する。
func unescapeXMLText(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		semicolon := strings.IndexByte(s[i:], ';')
		if semicolon < 0 {
			b.WriteByte(s[i])
			i++
			continue
		}
		entity := s[i : i+semicolon+1]
		var repl string
		var ok bool
		switch entity {
		case "&amp;":
			repl, ok = "&", true
		case "&lt;":
			repl, ok = "<", true
		case "&gt;":
			repl, ok = ">", true
		case "&apos;":
			repl, ok = "'", true
		case "&quot;":
			repl, ok = "\"", true
		}
		if ok {
			b.WriteString(repl)
			i += len(entity)
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
