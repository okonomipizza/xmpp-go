// Package jid は RFC 7622 に準拠した XMPP アドレス (JID) のパーサーを提供する。
//
// JID のフォーマット: [localpart@]domainpart[/resourcepart]
//
// 各パートの最大長は 1023 オクテット。
package jid

import (
	"errors"
	"strings"
)

// JID は XMPP アドレスを表す。
// RFC 7622 で定義されたフォーマット: [localpart@]domainpart[/resourcepart]
type JID struct {
	local    string
	domain   string
	resource string
}

// 各パートの最大長 (RFC 7622 Section 3.3, 3.4)
const maxPartLength = 1023

// localpart に含めてはいけない文字 (RFC 7622 Section 3.3.1)
// " & ' / : < > @
var localpartForbiddenChars = []rune{'"', '&', '\'', '/', ':', '<', '>', '@'}

var (
	// ErrEmptyJID は空の JID が渡された場合のエラー
	ErrEmptyJID = errors.New("jid: empty JID")

	// ErrEmptyDomain はドメインパートが空の場合のエラー
	ErrEmptyDomain = errors.New("jid: empty domain part")

	// ErrEmptyLocal はローカルパートが空 (@ があるが値がない) の場合のエラー
	ErrEmptyLocal = errors.New("jid: empty local part")

	// ErrEmptyResource はリソースパートが空 (/ があるが値がない) の場合のエラー
	ErrEmptyResource = errors.New("jid: empty resource part")

	// ErrLocalTooLong はローカルパートが 1023 オクテットを超えた場合のエラー
	ErrLocalTooLong = errors.New("jid: local part exceeds 1023 octets")

	// ErrDomainTooLong はドメインパートが 1023 オクテットを超えた場合のエラー
	ErrDomainTooLong = errors.New("jid: domain part exceeds 1023 octets")

	// ErrResourceTooLong はリソースパートが 1023 オクテットを超えた場合のエラー
	ErrResourceTooLong = errors.New("jid: resource part exceeds 1023 octets")

	// ErrForbiddenChar はローカルパートに禁止文字が含まれる場合のエラー
	ErrForbiddenChar = errors.New("jid: local part contains forbidden character")
)

// Parse は文字列から JID をパースする。
//
// RFC 7622 Section 3.1 に従い、以下の順序でパースする:
//  1. 最初の '/' 以降を resourcepart として分離
//  2. 最後の '@' より前を localpart として分離
//  3. 残りを domainpart とする
func Parse(s string) (JID, error) {
	if s == "" {
		return JID{}, ErrEmptyJID
	}

	var local, domain, resource string
	hasAt := false
	hasSlash := false

	// resourcepart の分離: 最初の '/' で分割
	if idx := strings.Index(s, "/"); idx != -1 {
		hasSlash = true
		resource = s[idx+1:]
		s = s[:idx]
	}

	// localpart の分離: 最後の '@' で分割
	// domainpart にも '@' が含まれる可能性があるため、最後の '@' を使用
	if idx := strings.LastIndex(s, "@"); idx != -1 {
		hasAt = true
		local = s[:idx]
		domain = s[idx+1:]
	} else {
		domain = s
	}

	// バリデーション

	// ドメインパートのバリデーション
	if domain == "" {
		return JID{}, ErrEmptyDomain
	}
	if len(domain) > maxPartLength {
		return JID{}, ErrDomainTooLong
	}

	// ローカルパートのバリデーション
	if hasAt && local == "" {
		return JID{}, ErrEmptyLocal
	}
	if local != "" {
		if len(local) > maxPartLength {
			return JID{}, ErrLocalTooLong
		}
		for _, forbidden := range localpartForbiddenChars {
			if strings.ContainsRune(local, forbidden) {
				return JID{}, ErrForbiddenChar
			}
		}
	}

	// リソースパートのバリデーション
	if hasSlash && resource == "" {
		return JID{}, ErrEmptyResource
	}
	if resource != "" && len(resource) > maxPartLength {
		return JID{}, ErrResourceTooLong
	}

	return JID{
		local:    local,
		domain:   domain,
		resource: resource,
	}, nil
}

// Local はローカルパートを返す。
func (j JID) Local() string {
	return j.local
}

// Domain はドメインパートを返す。
func (j JID) Domain() string {
	return j.domain
}

// Resource はリソースパートを返す。
func (j JID) Resource() string {
	return j.resource
}

// Bare はリソースパートを除いた JID (bare JID) を返す。
func (j JID) Bare() JID {
	return JID{
		local:  j.local,
		domain: j.domain,
	}
}

// String は JID を文字列に変換する。
func (j JID) String() string {
	var b strings.Builder

	if j.local != "" {
		b.WriteString(j.local)
		b.WriteByte('@')
	}
	b.WriteString(j.domain)
	if j.resource != "" {
		b.WriteByte('/')
		b.WriteString(j.resource)
	}

	return b.String()
}

// IsEmpty は JID が空かどうかを返す。
func (j JID) IsEmpty() bool {
	return j.domain == ""
}

// IsBare はリソースパートがない JID かどうかを返す。
func (j JID) IsBare() bool {
	return j.resource == ""
}

// IsFull はリソースパートがある JID かどうかを返す。
func (j JID) IsFull() bool {
	return j.resource != ""
}

// Equal は二つの JID が等しいかどうかを返す。
// 大文字小文字を区別する。
func (j JID) Equal(other JID) bool {
	return j.local == other.local &&
		j.domain == other.domain &&
		j.resource == other.resource
}
