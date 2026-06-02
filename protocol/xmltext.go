package protocol

import "bytes"

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
