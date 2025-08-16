package chat

import (
	"context"
	"fmt"
	"gochat/internal/util"
	"net"
	"strings"
	"sync"
    "github.com/google/uuid"
)

type Peer struct {
    uuid string // Unique identifier for the Peer
    Name string        // Name of the Peer
    Conn net.Conn    // Connection to the Peer
    mu sync.Mutex // Mutex to protect concurrent access to the connection
}


type ChatRoom struct {
    Peers []Peer    // List of connected peers
    mu sync.Mutex // Mutex to protect concurrent access to the chat room
}

func NewRoom() *ChatRoom {
    return &ChatRoom{
        Peers: make([]Peer, 0),
    }
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
    fmt.Println(util.Info, "Added peer:", name, "with UUID:", peer.uuid)
}

func PeerHandler(ctx context.Context, conn net.Conn, name string, room *ChatRoom) {
    defer conn.Close()
    fmt.Println(util.Info, "Handling connection from", conn.RemoteAddr())

    // Send and receive name for initial handshake
    if err := SendMsg(conn, name); err != nil {
        fmt.Println(util.Error, "Failed to send name:", err)
        return
    }
    
    receivedName, err := ReceiveMsg(conn)
    if err != nil {
        if err.Error() == "EOF" {
            fmt.Println(util.Info, "Peer closed connection during handshake")
        } else {
            fmt.Println(util.Error, "Failed to receive name:", err)
        }
        return
    }

    // Add peer to chat room
    room.AddPeer(receivedName, conn)

    fmt.Println(util.Info, "Peer", receivedName, "added to chat room")

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
                    // Connection closed by peer
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
            fmt.Println(util.Info, "PeerHandler shutting down for", conn.RemoteAddr())
            p := room.FindPeerByConn(conn)
            if p != nil {
                room.RemovePeer(p.uuid)
            }
            return
        case msg := <-messageChan:
            if msg == "" {
                continue
            }
            // Remove any trailing newlines and add exactly one
            msg = strings.TrimSpace(msg)
            if msg != "" {
                fmt.Println(msg)
            }
        case err := <-errorChan:
            if err != nil {
                if err.Error() == "connection closed by peer" {
                    fmt.Println(util.Info, "Peer closed connection:", conn.RemoteAddr())
                } else {
                    fmt.Println(util.Error, "Connection error:", err)
                }
                p := room.FindPeerByConn(conn)
                if p != nil {
                    room.RemovePeer(p.uuid)
                }
                return
            }
        }
    }
}

func SendMsg(conn net.Conn, msg string) error {
    // Remove any existing newlines and add exactly one
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


func Broadcast(room *ChatRoom, msg string){
    room.mu.Lock()
    defer room.mu.Unlock()

    for _, p := range room.Peers {
        p.mu.Lock()
        _ = SendMsg(p.Conn, msg)
        p.mu.Unlock()
    }
}

func (room *ChatRoom) RemovePeer( uuid string) {
    room.mu.Lock()
    defer room.mu.Unlock()

    for i, p := range room.Peers {
        if p.uuid == uuid {
            fmt.Println(util.Info, "Removing peer:", p.Name, "with UUID:", p.uuid)
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
    
    fmt.Println(util.Info, "Shutting down chat room, closing", len(cr.Peers), "connections")
    
    // Close all peer connections
    for _, peer := range cr.Peers {
        peer.Conn.Close()
    }
    
    // Clear the peers slice
    cr.Peers = cr.Peers[:0]
}
