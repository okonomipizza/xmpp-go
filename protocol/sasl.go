package protocol

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/okonomipizza/xmpp-go/xmlstream"
)

const saslNS = "urn:ietf:params:xml:ns:xmpp-sasl"

// AuthElementBytes は RFC 6120 Section 6.4 の <auth/> 要素を返す。
func AuthElementBytes(mechanism string, initial []byte) ([]byte, error) {
	if err := validateSASLMechanism(mechanism); err != nil {
		return nil, err
	}
	var body string
	if len(initial) > 0 {
		body = base64.StdEncoding.EncodeToString(initial)
	}
	return []byte(fmt.Sprintf(
		"<auth xmlns='%s' mechanism='%s'>%s</auth>",
		saslNS, mechanism, body,
	)), nil
}

// ResponseElementBytes は <response/> 要素を返す。
func ResponseElementBytes(payload []byte) []byte {
	body := base64.StdEncoding.EncodeToString(payload)
	return []byte(fmt.Sprintf("<response xmlns='%s'>%s</response>", saslNS, body))
}

// PLAINInitialResponse は SASL PLAIN の初期応答バイト列を返す (RFC 6120 Section 6.3.7)。
func PLAINInitialResponse(authz, username, password string) []byte {
	return []byte(authz + "\x00" + username + "\x00" + password)
}

func validateSASLMechanism(m string) error {
	if m == "" {
		return errors.New("protocol: empty SASL mechanism")
	}
	for _, r := range m {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '+' || r == '_' {
			continue
		}
		return fmt.Errorf("protocol: invalid SASL mechanism %q", m)
	}
	return nil
}

func elementSASLNamespace(tok xmlstream.Token) bool {
	return tok.AttrValue("xmlns") == saslNS
}

func (f StreamFeatures) hasMechanism(name string) bool {
	for _, m := range f.Mechanisms {
		if m == name {
			return true
		}
	}
	return false
}

func (c *Connection) selectMechanism() (string, error) {
	prefs := c.cfg.SASLMechanisms
	if len(prefs) == 0 {
		prefs = []string{"PLAIN"}
	}
	for _, p := range prefs {
		if c.features.hasMechanism(p) {
			return p, nil
		}
	}
	return "", errors.New("protocol: no supported SASL mechanism")
}

func saslInitial(cfg Config, mechanism string) ([]byte, error) {
	switch mechanism {
	case "PLAIN":
		if cfg.Password == "" {
			return nil, errors.New("protocol: password required for PLAIN")
		}
		user := cfg.JID.Local()
		if user == "" {
			return nil, errors.New("protocol: JID localpart required for PLAIN")
		}
		return PLAINInitialResponse("", user, cfg.Password), nil
	default:
		return nil, fmt.Errorf("protocol: no built-in initial response for %s", mechanism)
	}
}
