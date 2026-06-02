# xmpp-go

Sans-I/O XMPP protocol implementation in Go.

## Design Philosophy

This library implements the XMPP protocol as a **pure state machine** with no I/O operations. The core protocol logic is completely separated from network transport, making it:

- **Testable**: Property-based testing with arbitrary byte sequences
- **Portable**: Use with TCP, WebSocket, or any transport
- **Predictable**: No hidden side effects, deterministic behavior

## Installation

```bash
go get github.com/okonomipizza/xmpp-go
```

## Quick Start

```go
package main

import (
    "github.com/okonomipizza/xmpp-go/protocol"
)

func main() {
    // Create a connection state machine
    conn := protocol.NewConnection(protocol.Config{
        JID:      "user@example.com",
        Password: "secret",
    })

    // Feed bytes from network
    events, _ := conn.Receive(dataFromNetwork)

    // Process events
    for _, event := range events {
        switch e := event.(type) {
        case *protocol.MessageEvent:
            fmt.Printf("Message from %s: %s\n", e.From, e.Body)
        }
    }

    // Get bytes to send
    for {
        data := conn.BytesToSend()
        if data == nil {
            break
        }
        sendToNetwork(data)
    }
}
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Your Application                                   │
└─────────────────────────────────────────────────────┘
                    ↓ Events  ↑ Commands
┌─────────────────────────────────────────────────────┐
│  xmpp-go/protocol (this library)                    │
│  - Pure state machine                               │
│  - No I/O, no time.Now()                           │
│  - Input: []byte                                    │
│  - Output: []Event, []byte                          │
└─────────────────────────────────────────────────────┘
                    ↓ []byte  ↑ []byte
┌─────────────────────────────────────────────────────┐
│  Transport (your code)                              │
│  - TCP, TLS, WebSocket, etc.                        │
└─────────────────────────────────────────────────────┘
```

## RFC Compliance

- RFC 6120 - XMPP Core (`ref/rfc6120.txt`)
- RFC 6121 - XMPP IM (`ref/rfc6121.txt`)
- RFC 7622 - JID Format (`ref/rfc7622.txt`)

## Development

```bash
# Enter development environment
nix develop

# Run tests
go test ./...

# Run property-based tests
go test -v -run=Property ./...

# Static analysis
staticcheck ./...
```

## Status

**v0.x** - Under development, API unstable.

## License

MIT License
