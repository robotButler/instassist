package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type cliOption struct {
	name      string
	runPrompt func(ctx context.Context, prompt string) ([]byte, error)
}

const (
	defaultCLIName = "codex"
	titleText      = "instassist"

	helpInput   = "tab: switch cli • enter: send • shift+enter/alt+enter: newline"
	helpViewing = "enter: copy & exit • shift+enter/alt+enter: new prompt • j/k: scroll • ctrl-d/u: page • tab: switch cli • esc/q: quit"
)

type viewMode int

const (
	modeInput viewMode = iota
	modeRunning
	modeViewing
)

type exchange struct {
	cli      string
	prompt   string
	response string
	err      error
}

type responseMsg struct {
	output []byte
	err    error
	cli    string
}

type model struct {
	cliOptions []cliOption
	cliIndex   int

	input    textarea.Model
	viewport viewport.Model

	mode    viewMode
	running bool

	width  int
	height int
	ready  bool

	history      []exchange
	lastResponse string
	status       string
}

func main() {
	cliFlag := flag.String("cli", defaultCLIName, "default CLI to use: codex or claude")
	flag.Parse()

	m := newModel(*cliFlag)
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func newModel(defaultCLI string) model {
	cliOptions := []cliOption{
		{
			name: "codex",
			runPrompt: func(ctx context.Context, prompt string) ([]byte, error) {
				cmd := exec.CommandContext(ctx, "codex", "exec")
				cmd.Stdin = strings.NewReader(prompt)
				return cmd.CombinedOutput()
			},
		},
		{
			name: "claude",
			runPrompt: func(ctx context.Context, prompt string) ([]byte, error) {
				cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
				return cmd.CombinedOutput()
			},
		},
	}

	input := textarea.New()
	input.Placeholder = "Enter prompt"
	input.Focus()
	input.CharLimit = 0
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.SetHeight(5)

	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)

	cliIndex := 0
	for i, opt := range cliOptions {
		if strings.EqualFold(opt.name, defaultCLI) {
			cliIndex = i
			break
		}
	}

	return model{
		cliOptions: cliOptions,
		cliIndex:   cliIndex,
		input:      input,
		viewport:   viewport,
		mode:       modeInput,
		status:     helpInput,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeComponents()
		return m, nil
	case responseMsg:
		return m.handleResponse(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	default:
	}

	if m.mode == modeInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleResponse(msg responseMsg) (tea.Model, tea.Cmd) {
	m.running = false
	m.mode = modeViewing

	respText := strings.TrimRight(string(msg.output), "\n")
	if msg.err != nil && strings.TrimSpace(respText) == "" {
		respText = msg.err.Error()
	}
	ex := exchange{
		cli:      msg.cli,
		prompt:   m.input.Value(),
		response: respText,
		err:      msg.err,
	}
	m.history = append(m.history, ex)
	m.lastResponse = respText

	if msg.err != nil {
		m.status = fmt.Sprintf("error from %s • %s", msg.cli, helpViewing)
	} else {
		m.status = helpViewing
	}

	m.refreshViewportContent(true)
	m.resizeComponents()
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeInput:
		return m.handleInputKeys(msg)
	case modeRunning:
		return m.handleRunningKeys(msg)
	case modeViewing:
		return m.handleViewingKeys(msg)
	default:
		return m, nil
	}
}

func (m model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "tab" {
		m.toggleCLI()
		return m, nil
	}
	// Submit prompt on enter, add newline on shift+enter.
	if isShiftEnter(msg) {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return m, cmd
	}
	if msg.Type == tea.KeyEnter {
		return m.submitPrompt()
	}

	return m.updateInput(msg)
}

func (m model) handleViewingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc" || msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "tab":
		m.toggleCLI()
		return m, nil
	case isShiftEnter(msg):
		m.mode = modeInput
		m.running = false
		m.input.SetValue("")
		m.input.Focus()
		m.status = helpInput
		m.refreshViewportContent(false)
		m.resizeComponents()
		return m, nil
	case msg.Type == tea.KeyEnter:
		if m.lastResponse == "" {
			m.status = "nothing to copy • " + helpViewing
			return m, nil
		}
		if err := clipboard.WriteAll(m.lastResponse); err != nil {
			m.status = fmt.Sprintf("copy failed: %v • %s", err, helpViewing)
			return m, nil
		}
		return m, tea.Quit
	case msg.String() == "j":
		m.viewport.LineDown(1)
	case msg.String() == "k":
		m.viewport.LineUp(1)
	case msg.String() == "ctrl+d":
		m.viewport.HalfPageDown()
	case msg.String() == "ctrl+u":
		m.viewport.HalfPageUp()
	}
	return m, nil
}

func (m model) handleRunningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "tab" {
		m.toggleCLI()
		return m, nil
	}
	return m, nil
}

func (m model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) submitPrompt() (tea.Model, tea.Cmd) {
	prompt := strings.TrimRight(m.input.Value(), "\n")
	if strings.TrimSpace(prompt) == "" {
		m.status = "prompt is empty • " + helpInput
		return m, nil
	}

	m.running = true
	m.mode = modeRunning
	m.status = fmt.Sprintf("running %s… • tab: switch cli", m.currentCLI().name)

	selectedCLI := m.currentCLI()
	cliName := selectedCLI.name
	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		out, err := selectedCLI.runPrompt(ctx, prompt)
		return responseMsg{
			output: out,
			err:    err,
			cli:    cliName,
		}
	}

	m.resizeComponents()
	return m, cmd
}

func (m model) toggleCLI() {
	if len(m.cliOptions) == 0 {
		return
	}
	m.cliIndex = (m.cliIndex + 1) % len(m.cliOptions)
	if m.mode == modeInput {
		m.status = helpInput
	} else if m.mode == modeViewing {
		m.status = helpViewing
	}
}

func (m model) currentCLI() cliOption {
	return m.cliOptions[m.cliIndex]
}

func (m *model) resizeComponents() {
	if !m.ready {
		return
	}

	inputHeight := 0
	if m.mode != modeViewing {
		inputHeight = m.input.Height() + 1 // label line
	}

	helpLines := 1
	headerLines := 2
	spacing := 1
	viewportHeight := m.height - inputHeight - helpLines - headerLines - spacing
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	m.input.SetWidth(m.width - 2)
	m.viewport.Width = m.width - 2
	m.viewport.Height = viewportHeight

	m.refreshViewportContent(false)
}

func (m *model) refreshViewportContent(scrollToBottom bool) {
	var b strings.Builder
	for i, ex := range m.history {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("[%s] prompt:\n", ex.cli))
		b.WriteString(indent(ex.prompt, "  "))
		b.WriteString("\n")
		if ex.err != nil {
			b.WriteString("error:\n")
			b.WriteString(indent(strings.TrimSpace(ex.err.Error()), "  "))
		} else {
			b.WriteString("response:\n")
			b.WriteString(indent(ex.response, "  "))
		}
	}

	m.viewport.SetContent(b.String())
	if scrollToBottom {
		m.viewport.GotoBottom()
	}
}

func isShiftEnter(msg tea.KeyMsg) bool {
	if msg.Type == tea.KeyCtrlJ {
		return true
	}
	if msg.Type == tea.KeyEnter && msg.Alt {
		return true
	}
	s := msg.String()
	return s == "shift+enter" || s == "alt+enter"
}

func indent(s, prefix string) string {
	if s == "" {
		return prefix + "(empty)"
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func (m model) View() string {
	if !m.ready {
		return "loading..."
	}

	var b strings.Builder
	cli := m.currentCLI().name
	title := lipgloss.NewStyle().Bold(true).Render(titleText)
	b.WriteString(title)
	if m.running {
		b.WriteString("  ")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("running %s…", cli)))
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("CLI: %s (tab to switch)\n", cli))

	if len(m.history) > 0 {
		b.WriteString("\n")
		b.WriteString(m.viewport.View())
		b.WriteString("\n")
	}

	if m.mode != modeViewing {
		b.WriteString("\n")
		b.WriteString("Prompt:\n")
		b.WriteString(m.input.View())
		b.WriteString("\n")
	}

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.status)
		if m.mode == modeInput {
			b.WriteString(" • esc/q: cancel after response")
		}
	}

	return b.String()
}
