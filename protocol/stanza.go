package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/okonomipizza/xmpp-go/jid"
)

// AvailablePresenceOpts は RFC 6121 Section 4.2.1 の初期 presence 用オプション。
// 空文字 / nil は要素・属性を省略する。
type AvailablePresenceOpts struct {
	Lang     string
	Show     string
	Status   string
	Priority *int
}

// MessageStanzaBytes は RFC 6121 Section 5.2.1 形式の <message/> を返す。
func MessageStanzaBytes(from, to jid.JID, id, typ, lang, body string) ([]byte, error) {
	if err := validateStanzaFrom(from); err != nil {
		return nil, err
	}
	if to.IsEmpty() {
		return nil, errors.New("protocol: message to is required")
	}
	if err := validateStanzaAttr("id", id); err != nil {
		return nil, err
	}
	if err := validateStanzaAttr("type", typ); err != nil {
		return nil, err
	}
	if lang != "" {
		if err := validateXMLLang(lang); err != nil {
			return nil, err
		}
	}
	open := fmt.Sprintf("<message from='%s' to='%s' id='%s' type='%s'",
		from.String(), to.String(), id, typ)
	if lang != "" {
		open += fmt.Sprintf(" xml:lang='%s'", lang)
	}
	return []byte(open + "><body>" + escapeXMLText(body) + "</body></message>"), nil
}

// PresenceStanzaBytes は <presence/> を返す。to が empty なら to 属性なし。
func PresenceStanzaBytes(from, to jid.JID, id, typ string) ([]byte, error) {
	if err := validateStanzaFrom(from); err != nil {
		return nil, err
	}
	if id != "" {
		if err := validateStanzaAttr("id", id); err != nil {
			return nil, err
		}
	}
	if typ != "" {
		if err := validateStanzaAttr("type", typ); err != nil {
			return nil, err
		}
	}
	var attrs []string
	attrs = append(attrs, "from='"+from.String()+"'")
	if !to.IsEmpty() {
		attrs = append(attrs, "to='"+to.String()+"'")
	}
	if id != "" {
		attrs = append(attrs, "id='"+id+"'")
	}
	if typ != "" {
		attrs = append(attrs, "type='"+typ+"'")
	}
	return []byte("<presence " + strings.Join(attrs, " ") + "/>"), nil
}

// AvailablePresenceStanzaBytes は broadcast 初期 presence を返す (to / type なし)。
func AvailablePresenceStanzaBytes(from jid.JID, opts AvailablePresenceOpts) ([]byte, error) {
	if err := validateStanzaFrom(from); err != nil {
		return nil, err
	}
	if opts.Lang != "" {
		if err := validateXMLLang(opts.Lang); err != nil {
			return nil, err
		}
	}
	if err := validatePresenceShow(opts.Show); err != nil {
		return nil, err
	}
	if opts.Priority != nil {
		if err := validatePresencePriority(*opts.Priority); err != nil {
			return nil, err
		}
	}
	open := "<presence from='" + from.String() + "'"
	if opts.Lang != "" {
		open += " xml:lang='" + opts.Lang + "'"
	}
	var children []string
	if opts.Show != "" {
		children = append(children, "<show>"+escapeXMLText(opts.Show)+"</show>")
	}
	if opts.Status != "" {
		children = append(children, "<status>"+escapeXMLText(opts.Status)+"</status>")
	}
	if opts.Priority != nil {
		children = append(children, "<priority>"+strconv.Itoa(*opts.Priority)+"</priority>")
	}
	if len(children) == 0 {
		return []byte(open + "/>"), nil
	}
	return []byte(open + ">" + strings.Join(children, "") + "</presence>"), nil
}

// IQStanzaBytes は <iq/> を返す。inner は子要素の生 XML (検証は呼び出し側の責務)。
func IQStanzaBytes(from, to jid.JID, id, typ string, inner []byte) ([]byte, error) {
	if err := validateStanzaFrom(from); err != nil {
		return nil, err
	}
	if typ != "get" && typ != "set" && typ != "result" && typ != "error" {
		return nil, fmt.Errorf("protocol: invalid iq type %q", typ)
	}
	if err := validateStanzaAttr("id", id); err != nil {
		return nil, err
	}
	var attrs []string
	attrs = append(attrs, "from='"+from.String()+"'")
	if !to.IsEmpty() {
		attrs = append(attrs, "to='"+to.String()+"'")
	}
	attrs = append(attrs, "id='"+id+"'", "type='"+typ+"'")
	if len(inner) == 0 {
		return []byte("<iq " + strings.Join(attrs, " ") + "/>"), nil
	}
	out := append([]byte("<iq "+strings.Join(attrs, " ")+">"), inner...)
	out = append(out, []byte("</iq>")...)
	return out, nil
}

func validateStanzaFrom(from jid.JID) error {
	if from.IsEmpty() {
		return errors.New("protocol: stanza from is required")
	}
	if !from.IsFull() {
		return errors.New("protocol: stanza from must be a full JID")
	}
	return nil
}

func validateStanzaAttr(name, value string) error {
	if value == "" {
		return fmt.Errorf("protocol: empty stanza %s", name)
	}
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\'', '"', '<', '>', '&':
			return fmt.Errorf("protocol: stanza %s contains forbidden character", name)
		}
	}
	return nil
}

func validatePresenceShow(show string) error {
	if show == "" {
		return nil
	}
	switch show {
	case "away", "chat", "dnd", "xa":
		return nil
	default:
		return fmt.Errorf("protocol: invalid presence show %q", show)
	}
}

func validatePresencePriority(p int) error {
	if p < -128 || p > 127 {
		return fmt.Errorf("protocol: presence priority out of range")
	}
	return nil
}

