package xmlstream

// Kind はパーサーが返すトークンの種類を表す。
type Kind int

const (
	// KindPI は XML 宣言 (<?xml ...?>) を表す。
	KindPI Kind = iota
	// KindStreamOpen は未閉じの <stream:stream> 開始タグを表す。
	KindStreamOpen
	// KindStreamClose は </stream:stream> を表す。
	KindStreamClose
	// KindElement は完全な子要素 (開始タグから対応する終了タグまで) を表す。
	KindElement
)

// Attr は XML 属性を表す。
type Attr struct {
	Name  string
	Value string
}

// Token はパーサーが抽出した 1 つの論理単位を表す。
type Token struct {
	Kind  Kind
	Name  string // 例: "stream:stream", "message", "stream:features"
	Attrs []Attr
	Raw   []byte // 入力バッファ上の生バイト列 (コピーしない)
}

// AttrValue は属性名から値を返す。見つからなければ空文字列。
func (t Token) AttrValue(name string) string {
	for _, a := range t.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}
