package netx

import (
	"context"
	"fmt"
	"gochat/internal/chat"
	"gochat/internal/util"
	"net"
	"sync"
)


func Listen(port int) (net.Listener, error) {
    address := fmt.Sprintf(":%d", port);
    ln, error := net.Listen("tcp", address);
    if error != nil {
        fmt.Println(util.Error, "Failed to listen on port", port, ":", error)
        return nil, error
    }
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
            return
        default:
            conn, err := ln.Accept()
            if err != nil {
                fmt.Println(util.Error, "Failed to accept connection:", err)
                continue
            }
            
            connCtx, connCancel := context.WithCancel(ctx)
            
            go func() {
                defer connCancel()
                defer conn.Close()
                chat.PeerHandler(connCtx, conn, name, room)
            }()
        }
    }
}

