// Updated tui/tui.go with channel communication
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type errMsg error 

type Model struct {
    viewport viewport.Model
    textarea textarea.Model
    SenderStyle lipgloss.Style
    SystemStyle lipgloss.Style
    PeerStyle lipgloss.Style
    err error
    messages []string
    outgoingChan chan<- string // Channel to send outgoing messages
    incomingChan <-chan Message // Channel to receive incoming messages
}

type Message struct { 
    From string 
    Text string
}

type OutgoingMsg struct {
    Text string
}

// Original InitModel for backward compatibility
func InitModel() Model {
    return InitModelWithChannels(nil, nil)
}

// New function that accepts channels for message handling
func InitModelWithChannels(outgoingChan chan<- string, incomingChan <-chan Message) Model {
    ta := textarea.New()
    ta.Placeholder = "Type your message here..."
    ta.Focus()
    ta.Prompt = "> "
    ta.CharLimit = 512
    ta.SetWidth(40)
    ta.SetHeight(1)
    ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
    ta.ShowLineNumbers = false

    vp := viewport.New(40, 20)
    vp.SetContent("Welcome to the gochat application!\n\n")

    ta.KeyMap.InsertNewline.SetEnabled(false)

    return Model{
        viewport: vp,
        textarea: ta,
        SenderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
        SystemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
        PeerStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
        messages: []string{},
        err: nil,
        outgoingChan: outgoingChan,
        incomingChan: incomingChan,
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        textarea.Blink,
        listenForIncomingMessages(m.incomingChan),
    )
}

// Command to listen for incoming messages
func listenForIncomingMessages(incomingChan <-chan Message) tea.Cmd {
    if incomingChan == nil {
        return nil
    }
    return func() tea.Msg {
        select {
        case msg := <-incomingChan:
            return msg
        default:
            // Return nil to continue listening on next update
            return nil
        }
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var (
        tiCmd tea.Cmd
        vpCmd tea.Cmd
    )
    m.textarea, tiCmd = m.textarea.Update(msg)
    m.viewport, vpCmd = m.viewport.Update(msg)

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.viewport.Width = msg.Width
        m.textarea.SetWidth(msg.Width)
        m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
        if len(m.messages) > 0 {
            m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
        }
        m.viewport.GotoBottom()
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC, tea.KeyEsc:
            return m, tea.Quit
        case tea.KeyEnter:
            // Get the message text before resetting
            messageText := strings.TrimSpace(m.textarea.Value())
            if messageText == "" {
                return m, tea.Batch(tiCmd, vpCmd)
            }
            
            // Add to local display
            m.messages = append(m.messages, m.SenderStyle.Render("You: ")+messageText)
            m.viewport.SetContent(strings.Join(m.messages, "\n"))
            m.textarea.Reset()
            m.viewport.GotoBottom()
            
            // Send message through channel if available
            if m.outgoingChan != nil {
                go func() {
                    select {
                    case m.outgoingChan <- messageText:
                    default:
                        // Channel is full, message dropped
                    }
                }()
            }
            
            return m, tea.Batch(tiCmd, vpCmd, listenForIncomingMessages(m.incomingChan))
        }
    case Message: 
        // Format incoming messages with sender name and proper styling
        var formattedMsg string
        if msg.From == "System" {
            // System messages (join/leave notifications)
            formattedMsg = m.SystemStyle.Render(fmt.Sprintf("• %s", msg.Text))
        } else {
            // Peer messages - only color the name, not the entire message
            coloredName := m.PeerStyle.Render(msg.From)
            formattedMsg = fmt.Sprintf("%s: %s", coloredName, msg.Text)
        }
        m.messages = append(m.messages, formattedMsg)
        m.viewport.SetContent(strings.Join(m.messages, "\n"))
        m.viewport.GotoBottom()
        
        // Continue listening for more incoming messages
        return m, tea.Batch(tiCmd, vpCmd, listenForIncomingMessages(m.incomingChan))
        
    case errMsg:
        m.err = msg
        return m, nil
    }
    return m, tea.Batch(tiCmd, vpCmd, listenForIncomingMessages(m.incomingChan))
}

func (m Model) View() string {
    return fmt.Sprintf(
        "%s%s%s",
        m.viewport.View(),
        gap,
        m.textarea.View(),
    )
}
