package netx

import (
    "context"
    "fmt"
    "net"
    "sync"
    "gochat/internal/util"
    "gochat/internal/chat"
)


func Listen(port int) (net.Listener, error) {
    address := fmt.Sprintf(":%d", port);
    ln, error := net.Listen("tcp", address);
    if error != nil {
        fmt.Println(util.Error, "Failed to listen on port", port, ":", error)
        return nil, error
    }
    fmt.Println(util.Info, "Listening on port", port)
    return ln, nil
}

func AcceptConnections(ctx context.Context, ln net.Listener, name string, wg *sync.WaitGroup, room *chat.ChatRoom) {
    defer wg.Done()
    defer ln.Close()
    
    // Create a channel to signal when listener should stop
    done := make(chan struct{})
    
    // Start a goroutine to handle context cancellation
    go func() {
        <-ctx.Done()
        close(done)
    }()
    
    for {
        select {
        case <-done:
            fmt.Println(util.Info, "AcceptConnections shutting down...")
            return
        default:
            conn, err := ln.Accept()
            if err != nil {
                fmt.Println(util.Error, "Failed to accept connection:", err)
                continue
            }
            fmt.Println(util.Info, "Accepted connection from", conn.RemoteAddr())
            
            connCtx, connCancel := context.WithCancel(ctx)
            
            go func() {
                defer connCancel()
                defer conn.Close()
                chat.PeerHandler(connCtx, conn, name, room)
            }()
        }
    }
}

