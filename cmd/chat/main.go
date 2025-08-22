// Updated main.go with proper message handling
package main

import (
    "context"
    "fmt"
    "sync"
    "os"
    "gochat/internal/config"
    "gochat/internal/util"
    "gochat/internal/netx"
    "gochat/internal/chat"
    "gochat/internal/tui"
    tea "github.com/charmbracelet/bubbletea"
)

var flags config.Config

// Global channels for message handling
var (
    outgoingMsgChan = make(chan string, 100)
    incomingMsgChan = make(chan tui.Message, 100)
)

func main() {
    flags = config.Parse()
    var room = chat.NewRoom()
    var wg sync.WaitGroup

    // Set up the TUI message channel for the chat room
    room.SetTUIMessageChannel(incomingMsgChan)

    // Create TUI model with message channels
    model := tui.InitModelWithChannels(outgoingMsgChan, incomingMsgChan)
    p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
    
    ln, err := netx.Listen(flags.Port)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to listen on port %d: %v\n", flags.Port, err)
        os.Exit(1)
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    defer room.Shutdown()

    wg.Add(3)
    
    // Start network goroutines
    go netx.AcceptConnections(ctx, ln, flags.Name, &wg, room)
    go netx.DailPeers(ctx, flags.Peers, flags.Name, &wg, room)
    
    // Start TUI
    go func() {
        defer wg.Done()
        if _, err := p.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
            cancel()
            room.Shutdown()
        }
    }()
    
    // Handle messages between TUI and chat room
    go func() {
        defer wg.Done()
        for {
            select {
            case <-ctx.Done():
                return
            case msg := <-outgoingMsgChan:
                // Send outgoing message to all peers
                if msg != "" {
                    fullMsg := fmt.Sprintf("%s: %s", flags.Name, msg)
                    chat.Broadcast(room, fullMsg)
                }
            }
        }
    }()
    
    wg.Wait()
    fmt.Println(util.Info, "All goroutines finished, exiting...")
}

// Helper function to send messages to TUI from other parts of the application
func SendToTUI(from, text string) {
    select {
    case incomingMsgChan <- tui.Message{From: from, Text: text}:
    default:
        // Channel is full, drop message to prevent blocking
        fmt.Println(util.Warning, "TUI message channel full, dropping message")
    }
}
