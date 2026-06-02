package protocol

// StartTLSCommandBytes は RFC 6120 Section 5.4.2.1 の STARTTLS コマンドを返す。
func StartTLSCommandBytes() []byte {
	return []byte("<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")
}
