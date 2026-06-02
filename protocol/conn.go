// Package protocol は XMPP c2s の Sans I/O プロトコル状態機械を提供する。
// ネットワーク I/O は行わず、受信バイト列と送信バイト列のキューのみを扱う。
package protocol

import (
	"errors"
	"fmt"

	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// Connection はクライアント側 XMPP ストリームの状態機械である。
type Connection struct {
	cfg      Config
	state    State
	parser   *xmlstream.Parser
	out      [][]byte
	features StreamFeatures
}

// NewConnection は初期状態の接続を返す。
func NewConnection(cfg Config) *Connection {
	return &Connection{
		cfg:    cfg,
		state:  StateInitial,
		parser: xmlstream.NewParser(),
	}
}

// State は現在の交渉フェーズを返す。
func (c *Connection) State() State {
	return c.state
}

// NeedInput は次の Receive でより多くのバイトが必要なら true を返す。
func (c *Connection) NeedInput() bool {
	return c.parser.NeedInput()
}

// Start はクライアント初期ストリームヘッダを送信キューに載せ、状態を進める。
func (c *Connection) Start() error {
	if c.state != StateInitial {
		return errors.New("protocol: Start called more than once")
	}
	data, err := ClientStreamOpenBytes(c.cfg)
	if err != nil {
		return err
	}
	c.enqueue(data)
	c.state = StateAwaitServerStream
	return nil
}

// Receive は受信バイト列を状態機械に渡し、確定したイベントを返す。
// 入力が途中の場合はイベントなしで nil を返し、NeedInput() が true になる。
func (c *Connection) Receive(data []byte) ([]Event, error) {
	if c.state == StateClosed {
		return nil, errors.New("protocol: connection closed")
	}
	c.parser.Feed(data)

	var events []Event
	for {
		tok, err := c.parser.Next()
		if err == xmlstream.ErrNeedMore {
			return events, nil
		}
		if err != nil {
			return events, err
		}
		ev, err := c.dispatch(tok)
		if err != nil {
			return events, err
		}
		if ev != nil {
			events = append(events, ev)
		}
	}
}

// Features は直近で受信した <stream:features/> の解析結果を返す。
func (c *Connection) Features() StreamFeatures {
	return c.features
}

// StartTLS は STARTTLS コマンドを送信キューに載せる (RFC 6120 Section 5.4.2.1)。
func (c *Connection) StartTLS() error {
	if c.state != StateNegotiating {
		return fmt.Errorf("protocol: StartTLS in state %s", c.state)
	}
	if !c.features.StartTLSOffered {
		return errors.New("protocol: STARTTLS not offered")
	}
	c.enqueue(StartTLSCommandBytes())
	c.state = StateAwaitTLSProceed
	return nil
}

// ResetAfterTLS は TLS 確立後に XML ストリーム状態をリセットする (Section 5.3.2)。
// 呼び出し側が transport で TLS を完了した後に呼び、続けて Start() すること。
func (c *Connection) ResetAfterTLS() {
	c.parser.Reset()
	c.features = StreamFeatures{}
	c.state = StateInitial
}

// BytesToSend は送信キュー先頭のバイト列を取り出す。なければ nil。
func (c *Connection) BytesToSend() []byte {
	if len(c.out) == 0 {
		return nil
	}
	data := c.out[0]
	c.out = c.out[1:]
	return data
}

func (c *Connection) enqueue(data []byte) {
	c.out = append(c.out, data)
}

func (c *Connection) dispatch(tok xmlstream.Token) (Event, error) {
	switch tok.Kind {
	case xmlstream.KindPI:
		return nil, nil
	case xmlstream.KindStreamOpen:
		return c.handleStreamOpen(tok)
	case xmlstream.KindStreamClose:
		c.state = StateClosed
		return &StreamClosedEvent{}, nil
	case xmlstream.KindElement:
		return c.handleElement(tok)
	default:
		return nil, fmt.Errorf("protocol: unexpected token kind %v", tok.Kind)
	}
}

func (c *Connection) handleStreamOpen(tok xmlstream.Token) (Event, error) {
	switch c.state {
	case StateAwaitServerStream, StateNegotiating:
		from, err := parseStreamJID(tok.AttrValue("from"))
		if err != nil {
			return nil, err
		}
		to, err := parseStreamJID(tok.AttrValue("to"))
		if err != nil {
			return nil, err
		}
		c.state = StateAwaitFeatures
		return &StreamOpenedEvent{
			From:    from,
			To:      to,
			ID:      tok.AttrValue("id"),
			Version: tok.AttrValue("version"),
			Lang:    tok.AttrValue("xml:lang"),
		}, nil
	default:
		return nil, fmt.Errorf("protocol: unexpected stream open in state %s", c.state)
	}
}

func (c *Connection) handleElement(tok xmlstream.Token) (Event, error) {
	switch tok.Name {
	case "stream:features":
		switch c.state {
		case StateAwaitFeatures, StateNegotiating:
			c.state = StateNegotiating
			c.features = ParseStreamFeatures(tok)
			return &StreamFeaturesEvent{Token: tok}, nil
		default:
			return nil, fmt.Errorf("protocol: unexpected stream:features in state %s", c.state)
		}
	case "stream:error":
		c.state = StateClosed
		return &StreamErrorEvent{Token: tok}, nil
	case "proceed":
		if c.state == StateAwaitTLSProceed && elementTLSNamespace(tok) {
			return &StartTLSProceedEvent{}, nil
		}
		return nil, fmt.Errorf("protocol: unexpected proceed in state %s", c.state)
	case "failure":
		if c.state == StateAwaitTLSProceed && elementTLSNamespace(tok) {
			c.state = StateClosed
			return &StartTLSFailureEvent{Token: tok}, nil
		}
		if c.state == StateNegotiating {
			return nil, nil
		}
		return nil, fmt.Errorf("protocol: unexpected failure in state %s", c.state)
	case "message", "presence", "iq":
		if c.state == StateReady {
			return &StanzaEvent{Name: tok.Name, Token: tok}, nil
		}
		if c.state == StateNegotiating {
			return nil, nil
		}
		return nil, fmt.Errorf("protocol: unexpected stanza %q in state %s", tok.Name, c.state)
	default:
		if c.state == StateNegotiating {
			return nil, nil
		}
		if c.state == StateReady {
			return &ElementEvent{Name: tok.Name, Token: tok}, nil
		}
		return nil, fmt.Errorf("protocol: unexpected element %q in state %s", tok.Name, c.state)
	}
}

// SetReady はストリーム交渉完了後に呼び出し、stanza の送受信を有効にする。
func (c *Connection) SetReady() {
	if c.state == StateNegotiating || c.state == StateAwaitFeatures {
		c.state = StateReady
	}
}
