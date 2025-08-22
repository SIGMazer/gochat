# gochat

A peer-to-peer chat app in Go. It's terminal-based, uses the Bubble Tea TUI framework, and lets you connect to other peers directly without a central server.

## Features

- Chat with other peers over TCP
- Simple terminal UI (Bubble Tea)
- Join, leave, and message notifications
- No central server needed—just connect to peers by address
- Easy to run multiple instances for local testing

## Quick Start

### Build

```bash
git clone https://github.com/SIGMazer/gochat.git
cd gochat
go build -o gochat ./cmd/chat
```

### Run

Start a chat node:
```bash
./gochat -name Alice -port 9001
```

Connect to Alice from Bob:
```bash
./gochat -name Bob -port 9002 -peers 127.0.0.1:9001
```

You can add more peers by listing their addresses in `-peers`, separated by commas.

### Flags

- `-name` (required): Your chat handle
- `-port`: Port to listen on (default 9000)
- `-peers`: Comma-list of host:port for other peers

Example:
```bash
./gochat -name Carol -port 9003 -peers 127.0.0.1:9001,127.0.0.1:9002
```

## How it Works

- Each instance listens on a port and connects to any peers you give it
- Messages go over TCP and show up in the TUI
- The UI colors your name, peer names, and system messages differently
- If you quit, peers see a leave message
- All chat happens in your terminal

## Testing

Run all tests:
```bash
go test ./...
```

## Dependencies

- Bubble Tea (TUI)
- Lipgloss (styles)
- Google UUID

## License

MIT

---

Pull requests and issues are welcome. If something's broken or you want to add stuff, just open an issue or PR.
