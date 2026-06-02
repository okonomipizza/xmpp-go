// Package xmlstream は XMPP 向けの増分 XML パーサーを提供する (Sans I/O)。
// ネットワークやファイル I/O は行わず、呼び出し側がバイト列を供給する。
package xmlstream

import (
	"bytes"
	"strings"
)

const streamOpenName = "stream:stream"
const streamCloseName = "stream:stream"

// Parser は XMPP ストリーム向けの増分 XML パーサーである。
// ネットワーク I/O は行わず、呼び出し側が []byte を Feed する Sans I/O モデル。
type Parser struct {
	buf        []byte
	incomplete bool // 直前の Next が入力不足で失敗した
}

// NewParser は空のパーサーを返す。
func NewParser() *Parser {
	return &Parser{}
}

// Reset は内部バッファをクリアする。
func (p *Parser) Reset() {
	p.buf = p.buf[:0]
	p.incomplete = false
}

// NeedInput は次のトークンを組み立てるために追加の入力が必要なら true を返す。
func (p *Parser) NeedInput() bool {
	return p.incomplete
}

// Buffered は未処理バイト数を返す。
func (p *Parser) Buffered() int {
	return len(p.buf)
}

// Feed は受信バイト列を内部バッファに追加する。
func (p *Parser) Feed(data []byte) {
	p.buf = append(p.buf, data...)
	p.incomplete = false
}

// Next は次のトークンを返す。データが不足している場合は ErrNeedMore を返す。
func (p *Parser) Next() (Token, error) {
	p.incomplete = false
	for {
		if err := p.skipSpace(); err != nil {
			return Token{}, err
		}
		if len(p.buf) == 0 {
			return Token{}, p.needMore()
		}

		switch p.buf[0] {
		case '<':
			if len(p.buf) >= 2 && p.buf[1] == '!' {
				if err := p.checkForbiddenMarkup(); err != nil {
					return Token{}, err
				}
			}
			if len(p.buf) >= 4 && p.buf[1] == '!' && p.buf[2] == '-' && p.buf[3] == '-' {
				return Token{}, ErrRestrictedXML
			}
			if len(p.buf) >= 2 && p.buf[1] == '?' {
				return p.consumePI()
			}
			if len(p.buf) >= 2 && p.buf[1] == '/' {
				return p.consumeStreamClose()
			}
			if len(p.buf) >= 9 && string(p.buf[:9]) == "<![CDATA[" {
				return Token{}, ErrRestrictedXML
			}
			return p.consumeElement()
		default:
			return Token{}, ErrMalformed
		}
	}
}

func (p *Parser) skipSpace() error {
	i := 0
	for i < len(p.buf) && isXMLSpace(p.buf[i]) {
		i++
	}
	p.buf = p.buf[i:]
	return nil
}

func isXMLSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (p *Parser) consumePI() (Token, error) {
	end := bytes.Index(p.buf, []byte("?>"))
	if end < 0 {
		return Token{}, p.needMore()
	}
	raw := p.buf[:end+2]
	p.buf = p.buf[end+2:]
	return Token{Kind: KindPI, Name: "?xml", Raw: raw}, nil
}

func (p *Parser) consumeStreamClose() (Token, error) {
	if len(p.buf) < 2 || p.buf[0] != '<' || p.buf[1] != '/' {
		return Token{}, ErrMalformed
	}
	gt := bytes.IndexByte(p.buf, '>')
	if gt < 0 {
		return Token{}, p.needMore()
	}
	name, err := parseCloseTagName(p.buf[:gt+1])
	if err != nil {
		return Token{}, err
	}
	if name != streamCloseName {
		return Token{}, ErrMalformed
	}
	raw := p.buf[:gt+1]
	p.buf = p.buf[gt+1:]
	return Token{Kind: KindStreamClose, Name: streamCloseName, Raw: raw}, nil
}

func parseCloseTagName(tag []byte) (string, error) {
	// </name> または </name >
	if len(tag) < 4 || tag[0] != '<' || tag[1] != '/' {
		return "", ErrMalformed
	}
	inner := tag[2 : len(tag)-1]
	inner = bytes.TrimSpace(inner)
	return string(inner), nil
}

func (p *Parser) consumeElement() (Token, error) {
	start := 0
	openEnd, selfClosing, err := p.findOpenTagEnd(start)
	if err != nil {
		if err == ErrNeedMore {
			return Token{}, p.needMore()
		}
		return Token{}, err
	}
	openTag := p.buf[start : openEnd+1]
	name, attrs, err := parseOpenTag(openTag)
	if err != nil {
		return Token{}, err
	}

	if name == streamOpenName {
		raw := p.buf[:openEnd+1]
		p.buf = p.buf[openEnd+1:]
		return Token{
			Kind:  KindStreamOpen,
			Name:  name,
			Attrs: attrs,
			Raw:   raw,
		}, nil
	}

	if selfClosing {
		raw := p.buf[:openEnd+1]
		p.buf = p.buf[openEnd+1:]
		return Token{Kind: KindElement, Name: name, Attrs: attrs, Raw: raw}, nil
	}

	// 子要素: 対応する終了タグまで読む
	depth := 1
	names := []string{name}
	pos := openEnd + 1
	for depth > 0 {
		next, err := p.findNextMarkup(pos)
		if err != nil {
			return Token{}, err
		}
		pos = next
		markup, err := p.sliceFrom(pos)
		if err != nil {
			return Token{}, err
		}
		if len(markup) == 0 {
			return Token{}, p.needMore()
		}

		switch {
		case markup[0] != '<':
			return Token{}, ErrMalformed
		case len(markup) < 2:
			return Token{}, p.needMore()
		case markup[1] == '!':
			if len(markup) >= 9 && string(markup[:9]) == "<![CDATA[" {
				end := bytes.Index(p.buf[pos:], []byte("]]>"))
				if end < 0 {
					return Token{}, p.needMore()
				}
				pos += end + 3
				continue
			}
			return Token{}, ErrRestrictedXML
		case markup[1] == '/':
			closeEnd := bytes.IndexByte(p.buf[pos:], '>')
			if closeEnd < 0 {
				return Token{}, p.needMore()
			}
			closeName, err := parseCloseTagName(p.buf[pos : pos+closeEnd+1])
			if err != nil {
				return Token{}, err
			}
			if depth > len(names) || closeName != names[depth-1] {
				return Token{}, ErrMalformed
			}
			depth--
			names = names[:depth]
			pos += closeEnd + 1
		case markup[1] == '?':
			return Token{}, ErrRestrictedXML
		default:
			innerEnd, innerSelfClosing, err := p.findOpenTagEnd(pos)
			if err != nil {
				if err == ErrNeedMore {
					return Token{}, p.needMore()
				}
				return Token{}, err
			}
			if innerSelfClosing {
				pos = innerEnd + 1
			} else {
				innerName, _, err := parseOpenTag(p.buf[pos : innerEnd+1])
				if err != nil {
					return Token{}, err
				}
				if innerName == streamOpenName {
					return Token{}, ErrMalformed
				}
				depth++
				names = append(names, innerName)
				pos = innerEnd + 1
			}
		}
	}

	raw := p.buf[:pos]
	p.buf = p.buf[pos:]
	return Token{Kind: KindElement, Name: name, Attrs: attrs, Raw: raw}, nil
}

func (p *Parser) sliceFrom(pos int) ([]byte, error) {
	if pos >= len(p.buf) {
		return nil, p.needMore()
	}
	return p.buf[pos:], nil
}

func (p *Parser) findNextMarkup(pos int) (int, error) {
	for pos < len(p.buf) {
		if p.buf[pos] == '<' {
			return pos, nil
		}
		pos++
	}
	return 0, p.needMore()
}

// findOpenTagEnd は pos から始まる開始タグの終端インデックス ('>' の位置) を返す。
// selfClosing は /> で終わったかどうか。
func (p *Parser) findOpenTagEnd(pos int) (end int, selfClosing bool, err error) {
	if pos >= len(p.buf) || p.buf[pos] != '<' {
		return 0, false, ErrMalformed
	}
	inQuote := byte(0)
	i := pos + 1
	for i < len(p.buf) {
		c := p.buf[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			i++
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '>':
			if i > 0 && p.buf[i-1] == '/' {
				return i, true, nil
			}
			return i, false, nil
		}
		i++
	}
	return 0, false, ErrNeedMore
}

func parseOpenTag(tag []byte) (name string, attrs []Attr, err error) {
	if len(tag) < 2 || tag[0] != '<' {
		return "", nil, ErrMalformed
	}
	inner := tag[1:]
	if inner[len(inner)-1] == '>' {
		inner = inner[:len(inner)-1]
	}
	if len(inner) > 0 && inner[len(inner)-1] == '/' {
		inner = inner[:len(inner)-1]
	}
	inner = bytes.TrimSpace(inner)
	if len(inner) == 0 {
		return "", nil, ErrMalformed
	}

	// 名前と属性を分離
	parts := splitNameAndAttrs(inner)
	name = string(bytes.TrimSpace(parts.name))
	if name == "" {
		return "", nil, ErrMalformed
	}
	attrs, err = parseAttrs(parts.attrBytes)
	if err != nil {
		return "", nil, err
	}
	return name, attrs, nil
}

type nameAttrSplit struct {
	name      []byte
	attrBytes []byte
}

func splitNameAndAttrs(inner []byte) nameAttrSplit {
	// 最初の空白で名前と属性を分ける (引用符内は無視)
	inQuote := byte(0)
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = c
			continue
		}
		if isXMLSpace(c) {
			return nameAttrSplit{
				name:      inner[:i],
				attrBytes: bytes.TrimSpace(inner[i:]),
			}
		}
	}
	return nameAttrSplit{name: inner, attrBytes: nil}
}

func parseAttrs(b []byte) ([]Attr, error) {
	var attrs []Attr
	for len(b) > 0 {
		b = bytes.TrimSpace(b)
		if len(b) == 0 {
			break
		}
		eq := bytes.IndexByte(b, '=')
		if eq < 0 {
			return nil, ErrMalformed
		}
		name := strings.TrimSpace(string(b[:eq]))
		b = bytes.TrimSpace(b[eq+1:])
		if len(b) == 0 {
			return nil, ErrMalformed
		}
		quote := b[0]
		if quote != '"' && quote != '\'' {
			return nil, ErrMalformed
		}
		b = b[1:]
		var val []byte
		for i := 0; i < len(b); i++ {
			if b[i] == quote {
				val = b[:i]
				b = b[i+1:]
				break
			}
		}
		if val == nil {
			return nil, ErrMalformed
		}
		attrs = append(attrs, Attr{Name: name, Value: string(val)})
	}
	return attrs, nil
}

func (p *Parser) needMore() error {
	p.incomplete = true
	return ErrNeedMore
}

// checkForbiddenMarkup は <! で始まる禁止構造 (DOCTYPE など) を検査する。
func (p *Parser) checkForbiddenMarkup() error {
	if len(p.buf) >= 9 && string(p.buf[:9]) == "<![CDATA[" {
		return nil
	}
	if len(p.buf) >= 4 && p.buf[2] == '-' && p.buf[3] == '-' {
		return nil // コメントは別途 ErrRestrictedXML
	}
	// トークンが未完了の可能性
	if len(p.buf) < 4 {
		return p.needMore()
	}
	return ErrRestrictedXML
}
