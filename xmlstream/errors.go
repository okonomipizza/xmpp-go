package xmlstream

import "errors"

var (
	// ErrNeedMore は 1 つの完全なトークンを組み立てるのにバイトが足りない場合のエラー。
	ErrNeedMore = errors.New("xmlstream: need more data")

	// ErrRestrictedXML は XMPP で禁止された XML 構造 (コメント, DTD など) を検出した場合のエラー。
	ErrRestrictedXML = errors.New("xmlstream: restricted XML")

	// ErrMalformed は整形式でない XML を検出した場合のエラー。
	ErrMalformed = errors.New("xmlstream: malformed XML")
)
