// Package protocol は XMPP c2s の Sans I/O プロトコル状態機械を提供する。
// ネットワーク I/O は行わず、受信バイト列と送信バイト列のキューのみを扱う。
package protocol

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// Connection はクライアント側 XMPP ストリームの状態機械である。
type Connection struct {
	cfg           Config
	state         State
	parser        *xmlstream.Parser
	out           [][]byte
	features      StreamFeatures
	pendingBindID string
	boundJID      jid.JID
	bindIQCounter uint64
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

// Authenticate は SASL <auth/> を送信キューに載せる (RFC 6120 Section 6.4)。
// mechanism が空なら Config の優先順と features の交差から選ぶ。PLAIN の初期応答を内蔵する。
func (c *Connection) Authenticate(mechanism string) error {
	if c.state != StateNegotiating {
		return fmt.Errorf("protocol: Authenticate in state %s", c.state)
	}
	if len(c.features.Mechanisms) == 0 {
		return errors.New("protocol: SASL not offered")
	}
	var err error
	if mechanism == "" {
		mechanism, err = c.selectMechanism()
		if err != nil {
			return err
		}
	}
	if !c.features.hasMechanism(mechanism) {
		return fmt.Errorf("protocol: mechanism %q not offered", mechanism)
	}
	initial, err := saslInitial(c.cfg, mechanism)
	if err != nil {
		return err
	}
	data, err := AuthElementBytes(mechanism, initial)
	if err != nil {
		return err
	}
	c.enqueue(data)
	c.state = StateAwaitSASLOutcome
	return nil
}

// SASLResponse は <response/> を送信キューに載せる。
func (c *Connection) SASLResponse(payload []byte) error {
	if c.state != StateAwaitSASLOutcome {
		return fmt.Errorf("protocol: SASLResponse in state %s", c.state)
	}
	c.enqueue(ResponseElementBytes(payload))
	return nil
}

// ResetAfterSASL は SASL 成功後に XML ストリーム状態をリセットする (Section 6.3.2)。
func (c *Connection) ResetAfterSASL() {
	c.parser.Reset()
	c.features = StreamFeatures{}
	c.pendingBindID = ""
	c.state = StateInitial
}

// BoundJID は bind 成功後の full JID を返す。未 bind なら empty。
func (c *Connection) BoundJID() jid.JID {
	return c.boundJID
}

// Bind は Config.Resource で resource binding IQ を送信キューに載せる (RFC 6120 Section 7)。
func (c *Connection) Bind() error {
	return c.BindResource(c.cfg.Resource)
}

// BindResource は指定 resourcepart で bind IQ を送信キューに載せる。空ならサーバー生成。
func (c *Connection) BindResource(resource string) error {
	if c.state != StateNegotiating {
		return fmt.Errorf("protocol: Bind in state %s", c.state)
	}
	if !c.features.BindOffered {
		return errors.New("protocol: resource binding not offered")
	}
	c.bindIQCounter++
	id := "bind" + strconv.FormatUint(c.bindIQCounter, 10)
	data, err := BindIQSetBytes(id, resource)
	if err != nil {
		return err
	}
	c.enqueue(data)
	c.pendingBindID = id
	c.state = StateAwaitBind
	return nil
}

func (c *Connection) sendStanza(data []byte) error {
	if c.state != StateReady {
		return fmt.Errorf("protocol: send stanza in state %s", c.state)
	}
	if c.boundJID.IsEmpty() {
		return errors.New("protocol: bound JID required to send stanza")
	}
	c.enqueue(data)
	return nil
}

// SendMessage は RFC 6121 形式の <message/> を送信キューに載せる。
func (c *Connection) SendMessage(to jid.JID, id, typ, lang, body string) error {
	data, err := MessageStanzaBytes(c.boundJID, to, id, typ, lang, body)
	if err != nil {
		return err
	}
	return c.sendStanza(data)
}

// SendPresence は <presence/> を送信キューに載せる。
func (c *Connection) SendPresence(to jid.JID, id, typ string) error {
	data, err := PresenceStanzaBytes(c.boundJID, to, id, typ)
	if err != nil {
		return err
	}
	return c.sendStanza(data)
}

// SendIQ は <iq/> を送信キューに載せる。
func (c *Connection) SendIQ(to jid.JID, id, typ string, inner []byte) error {
	data, err := IQStanzaBytes(c.boundJID, to, id, typ, inner)
	if err != nil {
		return err
	}
	return c.sendStanza(data)
}

// Close は RFC 6120 Section 4.4 の </stream:stream> を送信キューに載せる。
// 呼び出し後は outbound を送らず、サーバーからの close を Receive で処理する。
func (c *Connection) Close() error {
	switch c.state {
	case StateInitial:
		return errors.New("protocol: Close before Start")
	case StateClosed:
		return errors.New("protocol: connection closed")
	case StateClosing:
		return errors.New("protocol: stream close already sent")
	}
	c.enqueue(StreamCloseBytes())
	c.state = StateClosing
	return nil
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
		if c.state == StateAwaitSASLOutcome && elementSASLNamespace(tok) {
			c.state = StateNegotiating
			return &SASLFailureEvent{Token: tok}, nil
		}
		if c.state == StateNegotiating {
			return nil, nil
		}
		return nil, fmt.Errorf("protocol: unexpected failure in state %s", c.state)
	case "challenge":
		if c.state == StateAwaitSASLOutcome && elementSASLNamespace(tok) {
			return &SASLChallengeEvent{
				Payload: elementTextContent(tok.Raw),
				Token:   tok,
			}, nil
		}
		return nil, fmt.Errorf("protocol: unexpected challenge in state %s", c.state)
	case "success":
		if c.state == StateAwaitSASLOutcome && elementSASLNamespace(tok) {
			return &SASLSuccessEvent{
				Payload: elementTextContent(tok.Raw),
				Token:   tok,
			}, nil
		}
		return nil, fmt.Errorf("protocol: unexpected success in state %s", c.state)
	case "iq":
		return c.handleIQ(tok)
	case "message":
		if c.state == StateReady || c.state == StateClosing {
			ev, err := ParseInboundMessage(tok)
			if err != nil {
				return nil, err
			}
			return &ev, nil
		}
		if c.state == StateNegotiating || c.state == StateAwaitSASLOutcome || c.state == StateAwaitBind {
			return nil, nil
		}
		return nil, fmt.Errorf("protocol: unexpected stanza %q in state %s", tok.Name, c.state)
	case "presence":
		if c.state == StateReady || c.state == StateClosing {
			return &StanzaEvent{Name: tok.Name, Token: tok}, nil
		}
		if c.state == StateNegotiating || c.state == StateAwaitSASLOutcome || c.state == StateAwaitBind {
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

func (c *Connection) handleIQ(tok xmlstream.Token) (Event, error) {
	if c.state != StateAwaitBind {
		if c.state == StateReady || c.state == StateClosing {
			return &StanzaEvent{Name: tok.Name, Token: tok}, nil
		}
		if c.state == StateNegotiating || c.state == StateAwaitSASLOutcome {
			return nil, nil
		}
		return nil, fmt.Errorf("protocol: unexpected iq in state %s", c.state)
	}
	if tok.AttrValue("id") != c.pendingBindID {
		return nil, fmt.Errorf("protocol: unexpected bind iq id %q", tok.AttrValue("id"))
	}
	switch tok.AttrValue("type") {
	case "result":
		j, err := parseBindResultJID(tok.Raw)
		if err != nil {
			return nil, err
		}
		c.boundJID = j
		c.pendingBindID = ""
		c.state = StateReady
		return &BindSuccessEvent{JID: j, Token: tok}, nil
	case "error":
		c.pendingBindID = ""
		c.state = StateNegotiating
		return &BindFailureEvent{Token: tok}, nil
	default:
		return nil, fmt.Errorf("protocol: unexpected bind iq type %q", tok.AttrValue("type"))
	}
}

// SetReady はストリーム交渉完了後に呼び出し、stanza の送受信を有効にする。
// bind を使わないテスト用。本番フローでは Bind() 成功後に StateReady になる。
func (c *Connection) SetReady() {
	if c.state == StateNegotiating || c.state == StateAwaitFeatures {
		c.state = StateReady
	}
}
