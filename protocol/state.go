package protocol

import "strconv"

// State は c2s 接続の交渉フェーズを表す。
type State int

const (
	// StateInitial は接続開始前。Start でクライアントストリームヘッダを送る。
	StateInitial State = iota
	// StateAwaitServerStream はサーバーストリームヘッダ待ち。
	StateAwaitServerStream
	// StateAwaitFeatures は最初の <stream:features/> 待ち。
	StateAwaitFeatures
	// StateNegotiating はストリーム機能の交渉中 (TLS, SASL など)。
	StateNegotiating
	// StateAwaitTLSProceed は <proceed/> または <failure/> 待ち。
	StateAwaitTLSProceed
	// StateReady は stanza の送受信が可能な状態。
	StateReady
	// StateClosed はストリームが閉じられた状態。
	StateClosed
)

// String はデバッグ用の状態名を返す。
func (s State) String() string {
	switch s {
	case StateInitial:
		return "Initial"
	case StateAwaitServerStream:
		return "AwaitServerStream"
	case StateAwaitFeatures:
		return "AwaitFeatures"
	case StateNegotiating:
		return "Negotiating"
	case StateAwaitTLSProceed:
		return "AwaitTLSProceed"
	case StateReady:
		return "Ready"
	case StateClosed:
		return "Closed"
	default:
		return "State(" + strconv.Itoa(int(s)) + ")"
	}
}
