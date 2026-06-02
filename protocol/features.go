package protocol

import (
	"bytes"

	"github.com/okonomipizza/xmpp-go/xmlstream"
)

const tlsNS = "urn:ietf:params:xml:ns:xmpp-tls"

// StreamFeatures は <stream:features/> から抽出した交渉情報。
type StreamFeatures struct {
	StartTLSOffered  bool
	StartTLSRequired bool
	Mechanisms       []string
	BindOffered      bool
}

type featureChild struct {
	name string
	raw  []byte
}

// ParseStreamFeatures は features トークンの直下の子要素から交渉情報を解析する。
func ParseStreamFeatures(tok xmlstream.Token) StreamFeatures {
	var f StreamFeatures
	for _, ch := range parseDirectChildElements(tok.Raw) {
		switch ch.name {
		case "starttls":
			f.StartTLSOffered = true
			for _, sub := range parseDirectChildElements(ch.raw) {
				if sub.name == "required" {
					f.StartTLSRequired = true
				}
			}
		case "mechanisms":
			f.Mechanisms = parseMechanisms(ch.raw)
		case "bind":
			f.BindOffered = true
		}
	}
	return f
}

func elementTLSNamespace(tok xmlstream.Token) bool {
	return tok.AttrValue("xmlns") == tlsNS
}

// parseDirectChildElements は要素 raw の直下の子要素を返す。
func parseDirectChildElements(elemRaw []byte) []featureChild {
	inner := elementInnerBytes(elemRaw)
	if inner == nil {
		return nil
	}
	var out []featureChild
	i := 0
	for i < len(inner) {
		for i < len(inner) && inner[i] != '<' {
			i++
		}
		if i >= len(inner) {
			break
		}
		if i+1 < len(inner) && inner[i+1] == '/' {
			if j := bytes.IndexByte(inner[i:], '>'); j >= 0 {
				i += j + 1
			} else {
				break
			}
			continue
		}
		if i+1 < len(inner) && (inner[i+1] == '?' || inner[i+1] == '!') {
			i = skipNonElementMarkup(inner, i)
			continue
		}
		start := i
		openEnd, selfClosing, ok := findOpenTagEnd(inner, i)
		if !ok {
			break
		}
		name := openTagLocalName(inner[i : openEnd+1])
		if selfClosing {
			out = append(out, featureChild{name: name, raw: inner[start : openEnd+1]})
			i = openEnd + 1
			continue
		}
		depth := 1
		names := []string{name}
		j := openEnd + 1
		for depth > 0 && j < len(inner) {
			k := bytes.IndexByte(inner[j:], '<')
			if k < 0 {
				break
			}
			j += k
			if j+1 >= len(inner) {
				break
			}
			switch inner[j+1] {
			case '/':
				closeEnd := bytes.IndexByte(inner[j:], '>')
				if closeEnd < 0 {
					break
				}
				closeEnd += j
				closeName := closeTagLocalName(inner[j : closeEnd+1])
				if closeName != names[depth-1] {
					return out
				}
				depth--
				names = names[:depth]
				j = closeEnd + 1
				if depth == 0 {
					out = append(out, featureChild{name: name, raw: inner[start:j]})
					i = j
				}
			case '?', '!':
				j = skipNonElementMarkup(inner, j)
			default:
				nestedEnd, nestedSelf, ok := findOpenTagEnd(inner, j)
				if !ok {
					return out
				}
				nestedName := openTagLocalName(inner[j : nestedEnd+1])
				if nestedSelf {
					j = nestedEnd + 1
				} else {
					depth++
					names = append(names, nestedName)
					j = nestedEnd + 1
				}
			}
		}
	}
	return out
}

func elementInnerBytes(elemRaw []byte) []byte {
	openEnd, _, ok := findOpenTagEnd(elemRaw, 0)
	if !ok {
		return nil
	}
	outerName := openTagLocalName(elemRaw[:openEnd+1])
	closeTag := []byte("</" + outerName + ">")
	closeStart := bytes.LastIndex(elemRaw, closeTag)
	if closeStart < 0 || closeStart <= openEnd {
		return nil
	}
	return elemRaw[openEnd+1 : closeStart]
}

// findOpenTagEnd は開始タグの '>' 終端を返す。属性値内の > は無視する (xmlstream.Parser と同じ規則)。
func findOpenTagEnd(buf []byte, start int) (end int, selfClosing bool, ok bool) {
	if start >= len(buf) || buf[start] != '<' {
		return 0, false, false
	}
	inQuote := byte(0)
	for i := start + 1; i < len(buf); i++ {
		c := buf[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '>':
			if i > start+1 && buf[i-1] == '/' {
				return i, true, true
			}
			return i, false, true
		}
	}
	return 0, false, false
}

func openTagLocalName(openTag []byte) string {
	if len(openTag) < 2 || openTag[0] != '<' {
		return ""
	}
	i := 1
	var b []byte
	for i < len(openTag) && openTag[i] != '>' && openTag[i] != '/' && openTag[i] != ' ' && openTag[i] != '\t' && openTag[i] != '\n' && openTag[i] != '\r' {
		b = append(b, openTag[i])
		i++
	}
	return string(b)
}

func closeTagLocalName(closeTag []byte) string {
	if len(closeTag) < 3 || closeTag[0] != '<' || closeTag[1] != '/' {
		return ""
	}
	i := 2
	var b []byte
	for i < len(closeTag) && closeTag[i] != '>' {
		if closeTag[i] != ' ' && closeTag[i] != '\t' && closeTag[i] != '\n' && closeTag[i] != '\r' {
			b = append(b, closeTag[i])
		}
		i++
	}
	return string(b)
}

func skipNonElementMarkup(buf []byte, start int) int {
	if i := bytes.IndexByte(buf[start:], '>'); i >= 0 {
		return start + i + 1
	}
	return len(buf)
}

func parseMechanisms(mechanismsRaw []byte) []string {
	var out []string
	for _, ch := range parseDirectChildElements(mechanismsRaw) {
		if ch.name == "mechanism" {
			out = append(out, elementTextContent(ch.raw))
		}
	}
	return out
}
