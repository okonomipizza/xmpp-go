package protocol

import (
	"fmt"

	"github.com/okonomipizza/xmpp-go/jid"
)

const (
	clientNS  = "jabber:client"
	streamNS  = "http://etherx.jabber.org/streams"
	streamVer = "1.0"
)

// ClientStreamOpenBytes は RFC 6120 Section 4.2 に沿ったクライアント初期ストリームヘッダを返す。
func ClientStreamOpenBytes(cfg Config) ([]byte, error) {
	if cfg.JID.IsEmpty() {
		return nil, fmt.Errorf("protocol: empty JID")
	}
	if cfg.JID.IsFull() {
		return nil, fmt.Errorf("protocol: JID must be bare for stream open")
	}
	lang := cfg.Lang
	if lang == "" {
		lang = "en"
	}
	if err := validateXMLLang(lang); err != nil {
		return nil, err
	}
	from := cfg.JID.String()
	to := cfg.JID.Domain()
	// RFC 6120 の例に合わせ単一引用符を使用
	return []byte(fmt.Sprintf(
		"<?xml version='1.0'?>\n"+
			"<stream:stream from='%s' to='%s' version='%s' xml:lang='%s' xmlns='%s' xmlns:stream='%s'>",
		from, to, streamVer, lang, clientNS, streamNS,
	)), nil
}

// validateXMLLang は xml:lang 属性値に載せて安全な文字列か検証する。
func validateXMLLang(lang string) error {
	for i := 0; i < len(lang); i++ {
		switch lang[i] {
		case '\'', '"', '<', '>', '&':
			return fmt.Errorf("protocol: xml:lang contains forbidden character")
		}
	}
	return nil
}

func parseStreamJID(attr string) (jid.JID, error) {
	if attr == "" {
		return jid.JID{}, nil
	}
	return jid.Parse(attr)
}
