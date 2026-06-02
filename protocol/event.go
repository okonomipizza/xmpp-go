package protocol

import (
	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// Event はプロトコル状態機械が通知する出来事のマーカーインターフェース。
type Event interface {
	isEvent()
}

// StreamOpenedEvent はサーバー (またはピア) のストリームヘッダを受信したことを表す。
type StreamOpenedEvent struct {
	From    jid.JID
	To      jid.JID
	ID      string
	Version string
	Lang    string
}

func (*StreamOpenedEvent) isEvent() {}

// StreamFeaturesEvent は <stream:features/> を受信したことを表す。
type StreamFeaturesEvent struct {
	Token xmlstream.Token
}

func (*StreamFeaturesEvent) isEvent() {}

// StartTLSProceedEvent は <proceed/> を受信し TLS ハンドシェイ可能なことを表す。
// 呼び出し側が transport で TLS を完了したら ResetAfterTLS() の後に Start() すること。
type StartTLSProceedEvent struct{}

func (*StartTLSProceedEvent) isEvent() {}

// StartTLSFailureEvent は STARTTLS の <failure/> を受信したことを表す。
type StartTLSFailureEvent struct {
	Token xmlstream.Token
}

func (*StartTLSFailureEvent) isEvent() {}

// SASLChallengeEvent は <challenge/> を受信したことを表す。Payload は Base64 文字列。
type SASLChallengeEvent struct {
	Payload string
	Token   xmlstream.Token
}

func (*SASLChallengeEvent) isEvent() {}

// SASLSuccessEvent は <success/> を受信したことを表す。
// transport は変更不要。ResetAfterSASL() の後に Start() すること。
type SASLSuccessEvent struct {
	Payload string
	Token   xmlstream.Token
}

func (*SASLSuccessEvent) isEvent() {}

// SASLFailureEvent は SASL <failure/> を受信したことを表す。
type SASLFailureEvent struct {
	Token xmlstream.Token
}

func (*SASLFailureEvent) isEvent() {}

// StanzaEvent は <message/>, <presence/>, <iq/> のいずれかを受信したことを表す。
type StanzaEvent struct {
	Name  string // "message", "presence", "iq"
	Token xmlstream.Token
}

func (*StanzaEvent) isEvent() {}

// ElementEvent は Ready 状態で受信した、stanza 以外のトップレベル要素を表す。
type ElementEvent struct {
	Name  string
	Token xmlstream.Token
}

func (*ElementEvent) isEvent() {}

// StreamClosedEvent は </stream:stream> を受信したことを表す。
type StreamClosedEvent struct{}

func (*StreamClosedEvent) isEvent() {}

// StreamErrorEvent は <stream:error> を受信したことを表す。
type StreamErrorEvent struct {
	Token xmlstream.Token
}

func (*StreamErrorEvent) isEvent() {}
