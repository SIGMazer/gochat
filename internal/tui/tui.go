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
    senderStyle lipgloss.Style
    err error
    messages []string
}

func InitModel() Model {
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
        senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
        messages: []string{},
        err: nil,
    }
}

func (m Model) Init() tea.Cmd {
    return textarea.Blink
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
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

