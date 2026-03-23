package editor

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorMode int

const (
	NormalMode EditorMode = iota
	InsertMode
	VisualMode
	CommandMode
)

type Cursor struct {
	X int
	Y int
}

type Selection struct {
	Start Cursor
	End   Cursor
	Active bool
}

type Clipboard struct {
	Content string
	Type    ClipboardType
}

type ClipboardType int

const (
	CharClipboard ClipboardType = iota
	WordClipboard
	LineClipboard
)

type ViEditor struct {
	lines         []string
	mode          EditorMode
	cursor        Cursor
	selection     Selection
	clipboard     Clipboard
	searchPattern string
	searchMatches []Match
	currentMatch  int
	commandBuffer string
	filename      string
	modified      bool
	width         int
	height        int
	scrollX       int
	scrollY       int
}

type Match struct {
	Line int
	Start int
	End   int
}

type ViModel struct {
	editor *ViEditor
	quit   bool
}

var (
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	modeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	lineNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func NewViEditor(content string, filename string) *ViEditor {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &ViEditor{
		lines:     lines,
		mode:      NormalMode,
		cursor:    Cursor{X: 0, Y: 0},
		selection: Selection{Active: false},
		clipboard: Clipboard{Content: "", Type: LineClipboard},
		filename:  filename,
		modified:  false,
		width:     80,
		height:    24,
		scrollX:   0,
		scrollY:   0,
	}
}

func (e *ViEditor) GetContent() string {
	return strings.Join(e.lines, "\n")
}

func (e *ViEditor) IsModified() bool {
	return e.modified
}

func (e *ViEditor) GetMode() EditorMode {
	return e.mode
}

func (e *ViEditor) SetMode(mode EditorMode) {
	e.mode = mode
	if mode != VisualMode {
		e.selection.Active = false
	}
}

func (e *ViEditor) MoveCursor(dx, dy int) {
	e.cursor.X += dx
	e.cursor.Y += dy

	// Bounds checking
	if e.cursor.Y < 0 {
		e.cursor.Y = 0
	}
	if e.cursor.Y >= len(e.lines) {
		e.cursor.Y = len(e.lines) - 1
	}

	if e.cursor.X < 0 {
		e.cursor.X = 0
	}
	if e.cursor.X > len(e.lines[e.cursor.Y]) {
		e.cursor.X = len(e.lines[e.cursor.Y])
	}

	e.updateScroll()
}

func (e *ViEditor) updateScroll() {
	// Vertical scrolling
	if e.cursor.Y < e.scrollY {
		e.scrollY = e.cursor.Y
	}
	if e.cursor.Y >= e.scrollY+e.height-2 {
		e.scrollY = e.cursor.Y - e.height + 3
	}

	// Horizontal scrolling
	if e.cursor.X < e.scrollX {
		e.scrollX = e.cursor.X
	}
	if e.cursor.X >= e.scrollX+e.width-10 {
		e.scrollX = e.cursor.X - e.width + 11
	}
}

func (e *ViEditor) InsertChar(char rune) {
	if e.mode != InsertMode {
		return
	}

	line := e.lines[e.cursor.Y]
	e.lines[e.cursor.Y] = line[:e.cursor.X] + string(char) + line[e.cursor.X:]
	e.cursor.X++
	e.modified = true
	e.updateScroll()
}

func (e *ViEditor) DeleteChar() {
	if e.cursor.X >= len(e.lines[e.cursor.Y]) {
		return
	}

	line := e.lines[e.cursor.Y]
	e.lines[e.cursor.Y] = line[:e.cursor.X] + line[e.cursor.X+1:]
	e.modified = true
}

func (e *ViEditor) DeleteLine() {
	if len(e.lines) == 1 {
		e.lines[0] = ""
		e.cursor.X = 0
		return
	}

	e.lines = append(e.lines[:e.cursor.Y], e.lines[e.cursor.Y+1:]...)
	if e.cursor.Y >= len(e.lines) {
		e.cursor.Y = len(e.lines) - 1
	}
	e.cursor.X = 0
	e.modified = true
	e.updateScroll()
}

func (e *ViEditor) YankLine() {
	if e.cursor.Y < len(e.lines) {
		e.clipboard.Content = e.lines[e.cursor.Y]
		e.clipboard.Type = LineClipboard
	}
}

func (e *ViEditor) Paste(after bool) {
	if e.clipboard.Content == "" {
		return
	}

	if e.clipboard.Type == LineClipboard {
		if after && e.cursor.Y < len(e.lines)-1 {
			e.cursor.Y++
		}
		
		newLines := strings.Split(e.clipboard.Content, "\n")
		e.lines = append(e.lines[:e.cursor.Y], append(newLines, e.lines[e.cursor.Y:]...)...)
		e.cursor.X = 0
		e.modified = true
		e.updateScroll()
	}
}

func (e *ViEditor) Search(pattern string, forward bool) {
	e.searchPattern = pattern
	e.searchMatches = []Match{}
	e.currentMatch = 0

	if pattern == "" {
		return
	}

	for y, line := range e.lines {
		indices := allStringIndices(line, pattern)
		for _, x := range indices {
			e.searchMatches = append(e.searchMatches, Match{Line: y, Start: x, End: x + len(pattern)})
		}
	}

	if len(e.searchMatches) > 0 {
		if forward {
			e.currentMatch = 0
		} else {
			e.currentMatch = len(e.searchMatches) - 1
		}
		match := e.searchMatches[e.currentMatch]
		e.cursor.Y = match.Line
		e.cursor.X = match.Start
		e.updateScroll()
	}
}

func allStringIndices(s, substr string) []int {
	var indices []int
	start := 0
	for {
		idx := strings.Index(s[start:], substr)
		if idx == -1 {
			break
		}
		indices = append(indices, start+idx)
		start += idx + 1
	}
	return indices
}

func (e *ViEditor) NextMatch() {
	if len(e.searchMatches) == 0 {
		return
	}

	e.currentMatch = (e.currentMatch + 1) % len(e.searchMatches)
	match := e.searchMatches[e.currentMatch]
	e.cursor.Y = match.Line
	e.cursor.X = match.Start
	e.updateScroll()
}

func (e *ViEditor) PrevMatch() {
	if len(e.searchMatches) == 0 {
		return
	}

	e.currentMatch = (e.currentMatch - 1 + len(e.searchMatches)) % len(e.searchMatches)
	match := e.searchMatches[e.currentMatch]
	e.cursor.Y = match.Line
	e.cursor.X = match.Start
	e.updateScroll()
}

func (e *ViEditor) ExecuteCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "w":
		if len(parts) > 1 {
			e.filename = parts[1]
		}
		return "save"
	case "q":
		if e.modified {
			return "No write since last change"
		}
		return "quit"
	case "wq":
		return "save_quit"
	case "s":
		if len(parts) >= 3 {
			old, new := parts[1], parts[2]
			global := len(parts) > 3 && parts[3] == "g"
			return e.replace(old, new, global)
		}
		return "Usage: :s/old/new/[g]"
	default:
		return fmt.Sprintf("Unknown command: %s", parts[0])
	}
}

func (e *ViEditor) replace(old, new string, global bool) string {
	count := 0
	for i, line := range e.lines {
		if global {
			newLine := strings.ReplaceAll(line, old, new)
			if newLine != line {
				e.lines[i] = newLine
				count++
			}
		} else {
			if strings.Contains(line, old) {
				e.lines[i] = strings.Replace(line, old, new, 1)
				count++
				break
			}
		}
	}

	if count > 0 {
		e.modified = true
		return fmt.Sprintf("Replaced %d occurrence(s)", count)
	}
	return "Pattern not found"
}

// Bubble Tea Model methods
func (m ViModel) Init() tea.Cmd {
	return nil
}

func (m ViModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg), nil
	case tea.WindowSizeMsg:
		m.editor.width = msg.Width
		m.editor.height = msg.Height
		m.editor.updateScroll()
	}
	return m, nil
}

func (m ViModel) handleKey(msg tea.KeyMsg) tea.Model {
	e := m.editor

	switch e.mode {
	case NormalMode:
		return m.handleNormalMode(msg)
	case InsertMode:
		return m.handleInsertMode(msg)
	case VisualMode:
		return m.handleVisualMode(msg)
	case CommandMode:
		return m.handleCommandMode(msg)
	}

	return m
}

func (m ViModel) handleNormalMode(msg tea.KeyMsg) tea.Model {
	e := m.editor

	switch msg.Type {
	case tea.KeyEsc:
		// Already in normal mode
	case tea.KeyEnter:
		// Move to start of next line
		e.MoveCursor(0, 1)
	case tea.KeyBackspace:
		// Move left
		e.MoveCursor(-1, 0)
	case tea.KeyDelete:
		e.DeleteChar()
	case tea.KeyUp:
		e.MoveCursor(0, -1)
	case tea.KeyDown:
		e.MoveCursor(0, 1)
	case tea.KeyLeft:
		e.MoveCursor(-1, 0)
	case tea.KeyRight:
		e.MoveCursor(1, 0)
	case tea.KeyRunes:
		switch msg.Runes[0] {
		case 'h':
			e.MoveCursor(-1, 0)
		case 'j':
			e.MoveCursor(0, 1)
		case 'k':
			e.MoveCursor(0, -1)
		case 'l':
			e.MoveCursor(1, 0)
		case 'i':
			e.SetMode(InsertMode)
		case 'a':
			e.MoveCursor(1, 0)
			e.SetMode(InsertMode)
		case 'o':
			e.lines = append(e.lines[:e.cursor.Y+1], append([]string{""}, e.lines[e.cursor.Y+1:]...)...)
			e.cursor.Y++
			e.cursor.X = 0
			e.SetMode(InsertMode)
			e.modified = true
		case 'O':
			e.lines = append(e.lines[:e.cursor.Y], append([]string{""}, e.lines[e.cursor.Y:]...)...)
			e.cursor.X = 0
			e.SetMode(InsertMode)
			e.modified = true
		case 'x':
			e.DeleteChar()
		case 'd':
			if len(msg.Runes) > 1 && msg.Runes[1] == 'd' {
				e.YankLine()
				e.DeleteLine()
			}
		case 'y':
			if len(msg.Runes) > 1 && msg.Runes[1] == 'y' {
				e.YankLine()
			}
		case 'p':
			e.Paste(true)
		case 'P':
			e.Paste(false)
		case '/':
			e.SetMode(CommandMode)
			e.commandBuffer = "/"
		case '?':
			e.SetMode(CommandMode)
			e.commandBuffer = "?"
		case ':':
			e.SetMode(CommandMode)
			e.commandBuffer = ":"
		case 'n':
			e.NextMatch()
		case 'N':
			e.PrevMatch()
		}
	}

	return m
}

func (m ViModel) handleInsertMode(msg tea.KeyMsg) tea.Model {
	e := m.editor

	switch msg.Type {
	case tea.KeyEsc:
		e.SetMode(NormalMode)
		e.MoveCursor(-1, 0)
	case tea.KeyEnter:
		// Split line
		line := e.lines[e.cursor.Y]
		e.lines[e.cursor.Y] = line[:e.cursor.X]
		e.lines = append(e.lines[:e.cursor.Y+1], append([]string{line[e.cursor.X:]}, e.lines[e.cursor.Y+1:]...)...)
		e.cursor.Y++
		e.cursor.X = 0
		e.modified = true
	case tea.KeyBackspace:
		if e.cursor.X > 0 {
			line := e.lines[e.cursor.Y]
			e.lines[e.cursor.Y] = line[:e.cursor.X-1] + line[e.cursor.X:]
			e.cursor.X--
			e.modified = true
		}
	case tea.KeyDelete:
		e.DeleteChar()
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			e.InsertChar(r)
		}
	}

	return m
}

func (m ViModel) handleVisualMode(msg tea.KeyMsg) tea.Model {
	// TODO: Implement visual mode
	return m
}

func (m ViModel) handleCommandMode(msg tea.KeyMsg) tea.Model {
	e := m.editor

	switch msg.Type {
	case tea.KeyEsc:
		e.SetMode(NormalMode)
		e.commandBuffer = ""
	case tea.KeyEnter:
		cmd := e.commandBuffer[1:] // Remove :, /, or ?
		if e.commandBuffer[0] == ':' {
			result := e.ExecuteCommand(cmd)
			if result == "quit" {
				m.quit = true
			} else if result == "save_quit" {
				m.quit = true
			}
		} else if e.commandBuffer[0] == '/' {
			e.Search(cmd, true)
		} else if e.commandBuffer[0] == '?' {
			e.Search(cmd, false)
		}
		e.SetMode(NormalMode)
		e.commandBuffer = ""
	case tea.KeyBackspace:
		if len(e.commandBuffer) > 1 {
			e.commandBuffer = e.commandBuffer[:len(e.commandBuffer)-1]
		}
	case tea.KeyRunes:
		e.commandBuffer += string(msg.Runes[0])
	}

	return m
}

func (m ViModel) View() string {
	e := m.editor
	var output strings.Builder

	// Calculate visible lines
	startY := e.scrollY
	endY := startY + e.height - 2
	if endY > len(e.lines) {
		endY = len(e.lines)
	}

	// Render file content with line numbers
	for y := startY; y < endY; y++ {
		lineNum := fmt.Sprintf("%4d", y+1)
		output.WriteString(lineNumStyle.Render(lineNum) + " ")

		line := e.lines[y]
		if len(line) > e.scrollX {
			visibleLine := line[e.scrollX:]
			if len(visibleLine) > e.width-8 {
				visibleLine = visibleLine[:e.width-8]
			}
			output.WriteString(visibleLine)
		}
		output.WriteString("\n")
	}

	// Fill remaining space
	for y := endY; y < startY+e.height-2; y++ {
		output.WriteString("    ~\n")
	}

	// Status line
	modeStr := ""
	switch e.mode {
	case NormalMode:
		modeStr = "NORMAL"
	case InsertMode:
		modeStr = "INSERT"
	case VisualMode:
		modeStr = "VISUAL"
	case CommandMode:
		modeStr = "COMMAND"
	}

	status := modeStyle.Render(modeStr)
	if e.filename != "" {
		status += " " + e.filename
	}
	if e.modified {
		status += " [+]"
	}
	if len(e.searchMatches) > 0 {
		status += fmt.Sprintf(" (%d/%d)", e.currentMatch+1, len(e.searchMatches))
	}

	output.WriteString(statusStyle.Render(status) + "\n")

	// Command line
	if e.mode == CommandMode {
		output.WriteString(e.commandBuffer)
	}

	return output.String()
}

func StartViEditor(content, filename string) (string, bool, error) {
	editor := NewViEditor(content, filename)
	model := ViModel{
		editor: editor,
		quit:   false,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}

	viModel := finalModel.(ViModel)
	return viModel.editor.GetContent(), viModel.editor.IsModified(), nil
}
