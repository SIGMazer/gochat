package main

import (
    "context"
    "fmt"
    "sync"
    "os"
    "bufio"
    "gochat/internal/config"
    "gochat/internal/util"
    "gochat/internal/netx"
    "gochat/internal/chat"
    "gochat/internal/tui"
    tea "github.com/charmbracelet/bubbletea"
)


func main() {

    flags  := config.Parse();
    var room = chat.NewRoom()
    var wg sync.WaitGroup

    p := tea.NewProgram(tui.InitModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(util.Info, "TUI started successfully")



    fmt.Println(util.Info, flags)
    ln, err := netx.Listen(flags.Port)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to listen on port %d: %v\n", flags.Port, err)
        os.Exit(1)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    defer room.Shutdown()

    wg.Add(3)
    go netx.AcceptConnections(ctx, ln, flags.Name, &wg, room)
    go netx.DailPeers(ctx, flags.Peers, flags.Name, &wg, room)

    go func() {
        defer wg.Done()
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            msg := scanner.Text()
            if msg == ":exit" {
                fmt.Println(util.Info, "Exiting chat...")
                cancel() // Signal shutdown to other goroutines
                return
            }
            fullMsg := fmt.Sprintf("%s: %s", flags.Name, msg)
            chat.Broadcast(room, fullMsg)
        }
    }()
    
    wg.Wait()
    fmt.Println(util.Info, "All goroutines finished, exiting...")
}

