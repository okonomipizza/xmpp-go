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
