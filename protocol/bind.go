package protocol

import (
	"errors"
	"fmt"

	"github.com/okonomipizza/xmpp-go/jid"
)

const bindNS = "urn:ietf:params:xml:ns:xmpp-bind"

// BindIQSetBytes は RFC 6120 Section 7.6.1 / 7.7.1 の bind IQ (type set) を返す。
// resource が空ならサーバー生成、非空なら <resource/> を含める。
func BindIQSetBytes(id, resource string) ([]byte, error) {
	if err := validateIQID(id); err != nil {
		return nil, err
	}
	if resource == "" {
		return []byte(fmt.Sprintf(
			"<iq id='%s' type='set'><bind xmlns='%s'/></iq>",
			id, bindNS,
		)), nil
	}
	if err := validateBindResource(resource); err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(
		"<iq id='%s' type='set'><bind xmlns='%s'><resource>%s</resource></bind></iq>",
		id, bindNS, resource,
	)), nil
}

func validateIQID(id string) error {
	if id == "" {
		return errors.New("protocol: empty IQ id")
	}
	for _, r := range id {
		if r == '\'' || r == '"' || r == '<' || r == '>' || r == '&' {
			return fmt.Errorf("protocol: IQ id contains forbidden character")
		}
	}
	return nil
}

func validateBindResource(resource string) error {
	if resource == "" {
		return errors.New("protocol: empty bind resource")
	}
	for i := 0; i < len(resource); i++ {
		switch resource[i] {
		case '\'', '"', '<', '>', '&':
			return fmt.Errorf("protocol: bind resource contains forbidden character")
		}
	}
	// プレースホルダー JID で resourcepart の長さ・空を検証する
	_, err := jid.Parse("bind@" + "example.com" + "/" + resource)
	return err
}

// parseBindResultJID は bind 成功 IQ の <jid/> テキストから full JID を返す。
func parseBindResultJID(iqRaw []byte) (jid.JID, error) {
	for _, ch := range parseDirectChildElements(iqRaw) {
		if ch.name != "bind" {
			continue
		}
		for _, sub := range parseDirectChildElements(ch.raw) {
			if sub.name == "jid" {
				text := elementTextContent(sub.raw)
				if text == "" {
					return jid.JID{}, errors.New("protocol: empty bind jid")
				}
				return jid.Parse(text)
			}
		}
	}
	return jid.JID{}, errors.New("protocol: bind result missing jid")
}
