// Updated chat/chat.go - Modified PeerHandler function
package chat

import (
	"context"
	"fmt"
	"gochat/internal/util"
	"net"
	"strings"
	"sync"
    "github.com/google/uuid"
    "gochat/internal/tui"
)

type Peer struct {
    uuid string
    Name string
    Conn net.Conn
    mu sync.Mutex
}

type ChatRoom struct {
    Peers []Peer
    mu sync.Mutex
    // Add channel for sending messages to TUI
    tuiMsgChan chan<- tui.Message
}

func NewRoom() *ChatRoom {
    return &ChatRoom{
        Peers: make([]Peer, 0),
    }
}

// New function to set the TUI message channel
func (cr *ChatRoom) SetTUIMessageChannel(ch chan<- tui.Message) {
    cr.tuiMsgChan = ch
}

func (cr *ChatRoom) AddPeer(name string, conn net.Conn) {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    peer := Peer{
        uuid: uuid.NewString(),
        Name: name,
        Conn: conn,
    }
    cr.Peers = append(cr.Peers, peer)
}

// Modified PeerHandler to send messages through channel instead of directly to TUI
func PeerHandler(ctx context.Context, conn net.Conn, name string, room *ChatRoom) {
    defer conn.Close()

    // Send and receive name for initial handshake
    if err := SendMsg(conn, name); err != nil {
        fmt.Println(util.Error, "Failed to send name:", err)
        return
    }
    
    receivedName, err := ReceiveMsg(conn)
    if err != nil {
        if err.Error() == "EOF" {
        } else {
            fmt.Println(util.Error, "Failed to receive name:", err)
        }
        return
    }

    // Add peer to chat room
    room.AddPeer(receivedName, conn)

    // Send join notification to TUI through channel
    if room.tuiMsgChan != nil {
        select {
        case room.tuiMsgChan <- tui.Message{From: "System", Text: fmt.Sprintf("%s joined the chat", receivedName)}:
        default:
            fmt.Println(util.Warning, "TUI message channel full, dropping join notification")
        }
    }

    // Handle incoming messages with proper context handling
    messageChan := make(chan string, 1)
    errorChan := make(chan error, 1)
    
    // Start a goroutine to read messages
    go func() {
        defer close(messageChan)
        defer close(errorChan)
        
        for {
            msg, err := ReceiveMsg(conn)
            if err != nil {
                if err.Error() == "EOF" {
                    errorChan <- fmt.Errorf("connection closed by peer")
                } else {
                    errorChan <- err
                }
                return
            }
            messageChan <- msg
        }
    }()
    
    // Handle messages and context cancellation
    for {
        select {
        case <-ctx.Done():
            pe := room.FindPeerByConn(conn)
            if pe != nil {
                room.RemovePeer(pe.uuid)
                // Send leave notification to TUI through channel
                if room.tuiMsgChan != nil {
                    select {
                    case room.tuiMsgChan <- tui.Message{From: "System", Text: fmt.Sprintf("%s left the chat", receivedName)}:
                    default:
                        fmt.Println(util.Warning, "TUI message channel full, dropping leave notification")
                    }
                }
            }
            return
        case msg := <-messageChan:
            if msg == "" {
                continue
            }
            // Clean the message
            msg = strings.TrimSpace(msg)
            if msg != "" {
                // Parse the message to extract sender and content
                // Messages are formatted as "Sender: content"
                parts := strings.SplitN(msg, ":", 2)
                var sender, content string
                if len(parts) == 2 {
                    sender = strings.TrimSpace(parts[0])
                    content = strings.TrimSpace(parts[1])
                } else {
                    // Fallback if message doesn't have expected format
                    sender = receivedName
                    content = msg
                }
                
                // Send message to TUI through channel
                if room.tuiMsgChan != nil {
                    select {
                    case room.tuiMsgChan <- tui.Message{From: sender, Text: content}:
                    default:
                        fmt.Println(util.Warning, "TUI message channel full, dropping message from", sender)
                    }
                }
            }
        case err := <-errorChan:
            if err != nil {
                if err.Error() == "connection closed by peer" {
                } else {
                    fmt.Println(util.Error, "Connection error:", err)
                }
                peer := room.FindPeerByConn(conn)
                if peer != nil {
                    room.RemovePeer(peer.uuid)
                    // Send leave notification to TUI through channel
                    if room.tuiMsgChan != nil {
                        select {
                        case room.tuiMsgChan <- tui.Message{From: "System", Text: fmt.Sprintf("%s left the chat", receivedName)}:
                        default:
                            fmt.Println(util.Warning, "TUI message channel full, dropping leave notification")
                        }
                    }
                }
                return
            }
        }
    }
}

func SendMsg(conn net.Conn, msg string) error {
    msg = strings.TrimSpace(msg) + "\n"
    _, err := conn.Write([]byte(msg))
    if err != nil {
        fmt.Println(util.Error, "Failed to send message:", err)
        return err
    }
    return nil
}

func ReceiveMsg(conn net.Conn) (string, error) {
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        if err.Error() == "EOF" {
            return "", err
        }
        return "", fmt.Errorf("read error: %w", err)
    }
    if n == 0 {
        return "", fmt.Errorf("connection closed")
    }
    msg := string(buffer[:n])
    return msg, nil
}

func Broadcast(room *ChatRoom, msg string) {
    room.mu.Lock()
    defer room.mu.Unlock()

    for _, p := range room.Peers {
        p.mu.Lock()
        _ = SendMsg(p.Conn, msg)
        p.mu.Unlock()
    }
}

func (room *ChatRoom) RemovePeer(uuid string) {
    room.mu.Lock()
    defer room.mu.Unlock()

    for i, p := range room.Peers {
        if p.uuid == uuid {
            room.Peers = append(room.Peers[:i], room.Peers[i+1:]...)
            return
        }
    }
    fmt.Println(util.Warning, "Peer with UUID", uuid, "not found")
}

func (cr *ChatRoom) FindPeerByConn(conn net.Conn) *Peer {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    for i := range cr.Peers {
        if cr.Peers[i].Conn == conn {
            return &cr.Peers[i]
        }
    }
    return nil
}

func (cr *ChatRoom) Shutdown() {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    
    for _, peer := range cr.Peers {
        peer.Conn.Close()
    }
    
    cr.Peers = cr.Peers[:0]
}
