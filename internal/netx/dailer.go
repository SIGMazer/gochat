package netx 

import (
    "context"
    "fmt"
    "net"
    "sync"
    "time"
    "gochat/internal/util"
    "gochat/internal/chat"
)


func Dail(addr string) (net.Conn, error) {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        fmt.Println(util.Error, "Failed to connect to", addr, ":", err)
        return nil, err
    }
    fmt.Println(util.Info, "Connected to", addr)
    return conn, nil
}

func DailPeers(ctx context.Context, peers []string, name string, wg *sync.WaitGroup, room *chat.ChatRoom) {
    defer wg.Done()
    
    if len(peers) == 0 {
        return
    }
    
    for _, peer := range peers {
        select {
        case <-ctx.Done():
            fmt.Println(util.Info, "DailPeers shutting down...")
            return
        default:
            // Skip empty peer addresses
            if peer == "" {
                continue
            }
            
            conn, err := Dail(peer)
            if err != nil {
                fmt.Println(util.Warning, "Failed to connect to peer", peer, ":", err)
                continue
            }
            fmt.Println(util.Info, "Successfully connected to peer", peer)

            connCtx, connCancel := context.WithCancel(ctx)
            
            go func() {
                defer connCancel()
                defer conn.Close()
                chat.PeerHandler(connCtx, conn, name, room)
            }()
            
            // Small delay to prevent overwhelming the system
            time.Sleep(100 * time.Millisecond)
        }
    }
}

