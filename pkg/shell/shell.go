package shell

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/ops"
	"github.com/rmasci/k8sh/pkg/editor"
)

var (
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type Shell struct {
	client         *k8s.Client
	fileOps       *ops.FileOperations
	currentDir    string
	currentPod    string
	currentContainer string
	currentNamespace string
	history       []string
	input         string
}

type ShellModel struct {
	shell  *Shell
	output []string
}

func NewShell(client *k8s.Client) *Shell {
	fileOps := ops.NewFileOperations(client.GetClientset(), client.GetConfig())
	return &Shell{
		client:         client,
		fileOps:       fileOps,
		currentDir:    "/",
		currentPod:     "",
		currentContainer: "",
		currentNamespace: "default",
		history:       []string{},
		input:         "",
	}
}

func (s *Shell) ExecuteCommand(ctx context.Context, cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help":
		return s.showHelp()
	case "exit", "quit":
		return "exit"
	case "pwd":
		return s.currentDir
	case "ls":
		return s.listDirectory(ctx, args...)
	case "cd":
		return s.changeDirectory(args...)
	case "cat":
		return s.readFile(ctx, args...)
	case "vi", "vim":
		return s.editFile(ctx, args...)
	case "mkdir":
		return s.makeDirectory(ctx, args...)
	case "rm":
		return s.removeFile(ctx, args...)
	case "cp":
		return s.copyFile(ctx, args...)
	case "mv":
		return s.moveFile(ctx, args...)
	case "touch":
		return s.touchFile(ctx, args...)
	case "head":
		return s.headFile(ctx, args...)
	case "tail":
		return s.tailFile(ctx, args...)
	case "grep":
		return s.grepFile(ctx, args...)
	case "wc":
		return s.wordCount(ctx, args...)
	case "sort":
		return s.sortFile(ctx, args...)
	case "ps":
		return s.listProcesses(ctx)
	case "env":
		return s.getEnvironment(ctx)
	case "df":
		return s.getDiskUsage(ctx, args...)
	case "du":
		return s.getDirectoryUsage(ctx, args...)
	case "ip":
		return s.getPodIP(ctx)
	case "pods":
		return s.listPods(ctx)
	case "use":
		return s.selectPod(args...)
	case "namespace":
		return s.setNamespace(args...)
	case "clear":
		return "clear"
	default:
		return errorStyle.Render(fmt.Sprintf("Command not found: %s", command))
	}
}

func (s *Shell) showHelp() string {
	return `Available commands:
  help                    - Show this help message
  exit                    - Exit the shell
  pwd                     - Print current directory
  ls [path]              - List directory contents
  cd <path>              - Change directory
  cat <file>             - Display file contents
  vi/vim <file>          - Edit file with vi editor
  
  File Operations:
  mkdir <path>           - Create directory
  rm [-r] <path>         - Remove file/directory
  cp <src> <dst>         - Copy file
  mv <src> <dst>         - Move/rename file
  touch <file>           - Create empty file
  
  Text Processing:
  head [-n] <file>       - Show first lines
  tail [-n] <file>       - Show last lines
  grep [-i] <pattern> <file> - Search in file
  wc <file>              - Word count
  sort [-u] <file>       - Sort lines
  
  System Info:
  ps                      - List processes
  env                     - Environment variables
  df [path]              - Disk usage
  du [path]              - Directory usage
  ip                      - Show pod IP address
  
  Kubernetes:
  pods                   - List available pods
  use <pod> [container]  - Select pod and container
  namespace <name>       - Set current namespace
  
  Other:
  clear                  - Clear screen`
}

func (s *Shell) listPods(ctx context.Context) string {
	if s.currentPod == "" {
		pods, err := s.client.ListPods(ctx, s.currentNamespace)
		if err != nil {
			return errorStyle.Render(fmt.Sprintf("Error listing pods: %v", err))
		}

		if len(pods) == 0 {
			return "No pods found in namespace " + s.currentNamespace
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Pods in namespace '%s':\n", s.currentNamespace))
		for _, pod := range pods {
			result.WriteString(fmt.Sprintf("  %s (%s) - %s\n", pod.Name, pod.Status, pod.Node))
			for _, container := range pod.Containers {
				status := "Not Ready"
				if container.Ready {
					status = "Ready"
				}
				result.WriteString(fmt.Sprintf("    └─ %s (%s) - %s\n", container.Name, container.Image, status))
			}
		}
		return result.String()
	} else {
		return fmt.Sprintf("Currently using pod: %s/%s (container: %s)", s.currentNamespace, s.currentPod, s.currentContainer)
	}
}

func (s *Shell) selectPod(args ...string) string {
	if len(args) == 0 {
		s.currentPod = ""
		s.currentContainer = ""
		return "Pod selection cleared"
	}

	podName := args[0]
	containerName := ""
	if len(args) > 1 {
		containerName = args[1]
	}

	// Get pod info to find containers
	ctx := context.Background()
	pod, err := s.client.GetPod(ctx, s.currentNamespace, podName)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error getting pod: %v", err))
	}

	if containerName == "" {
		// Use first container if not specified
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			return errorStyle.Render("Pod has no containers")
		}
	}

	s.currentPod = podName
	s.currentContainer = containerName
	return fmt.Sprintf("Selected pod: %s/%s (container: %s)", s.currentNamespace, podName, containerName)
}

func (s *Shell) setNamespace(args ...string) string {
	if len(args) == 0 {
		return fmt.Sprintf("Current namespace: %s", s.currentNamespace)
	}

	s.currentNamespace = args[0]
	s.currentPod = ""
	s.currentContainer = ""
	return fmt.Sprintf("Namespace set to: %s", s.currentNamespace)
}

func (s *Shell) listDirectory(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	path := s.currentDir
	if len(args) > 0 {
		path = args[0]
	}

	files, err := s.fileOps.ListDirectory(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error listing directory: %v", err))
	}

	var result strings.Builder
	for _, file := range files {
		result.WriteString(file + "\n")
	}
	return result.String()
}

func (s *Shell) changeDirectory(args ...string) string {
	if len(args) == 0 {
		s.currentDir = "/"
		return ""
	}
	
	path := args[0]
	if path == ".." {
		if s.currentDir != "/" {
			s.currentDir = "/"
		}
	} else if strings.HasPrefix(path, "/") {
		s.currentDir = path
	} else {
		s.currentDir = s.currentDir + "/" + path
	}
	
	return ""
}

func (s *Shell) readFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: cat <file>"
	}
	
	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error reading file: %v", err))
	}

	return content
}

func (s *Shell) editFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: vi <file>"
	}
	
	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	// Read file content
	content, err := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error reading file: %v", err))
	}

	// Start vi editor
	editedContent, modified, err := editor.StartViEditor(content, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Editor error: %v", err))
	}

	// Save if modified
	if modified {
		err = s.fileOps.WriteFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path, editedContent)
		if err != nil {
			return errorStyle.Render(fmt.Sprintf("Error saving file: %v", err))
		}
		return fmt.Sprintf("File saved: %s", path)
	}

	return "File unchanged"
}

// Essential file tools
func (s *Shell) makeDirectory(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: mkdir <path>"
	}

	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	err := s.fileOps.MakeDirectory(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error creating directory: %v", err))
	}

	return fmt.Sprintf("Directory created: %s", path)
}

func (s *Shell) removeFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: rm [-r] <path>"
	}

	recursive := false
	path := args[0]
	if len(args) > 1 {
		if args[0] == "-r" {
			recursive = true
			path = args[1]
		}
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	err := s.fileOps.RemoveFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path, recursive)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error removing file: %v", err))
	}

	return fmt.Sprintf("Removed: %s", path)
}

func (s *Shell) copyFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) < 2 {
		return "Usage: cp <src> <dst>"
	}

	src, dst := args[0], args[1]
	if !strings.HasPrefix(src, "/") {
		src = s.currentDir + "/" + src
	}
	if !strings.HasPrefix(dst, "/") {
		dst = s.currentDir + "/" + dst
	}

	err := s.fileOps.CopyFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, src, dst)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error copying file: %v", err))
	}

	return fmt.Sprintf("Copied: %s -> %s", src, dst)
}

func (s *Shell) moveFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) < 2 {
		return "Usage: mv <src> <dst>"
	}

	src, dst := args[0], args[1]
	if !strings.HasPrefix(src, "/") {
		src = s.currentDir + "/" + src
	}
	if !strings.HasPrefix(dst, "/") {
		dst = s.currentDir + "/" + dst
	}

	err := s.fileOps.MoveFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, src, dst)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error moving file: %v", err))
	}

	return fmt.Sprintf("Moved: %s -> %s", src, dst)
}

func (s *Shell) touchFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: touch <file>"
	}

	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	err := s.fileOps.TouchFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error touching file: %v", err))
	}

	return fmt.Sprintf("File touched: %s", path)
}

// Text processing tools
func (s *Shell) headFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	lines := 10
	path := ""
	
	if len(args) == 1 {
		path = args[0]
	} else if len(args) == 2 && args[0] == "-n" {
		if num, err := strconv.Atoi(args[1]); err == nil {
			lines = num
		}
	} else if len(args) == 3 && args[0] == "-n" {
		if num, err := strconv.Atoi(args[1]); err == nil {
			lines = num
			path = args[2]
		}
	}

	if path == "" {
		return "Usage: head [-n] <file>"
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.HeadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path, lines)
	if err != nil {
		// Fallback to local processing
		fullContent, readErr := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
		if readErr != nil {
			return errorStyle.Render(fmt.Sprintf("Error reading file: %v", readErr))
		}
		content = ops.HeadLocal(fullContent, lines)
	}

	return content
}

func (s *Shell) tailFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	lines := 10
	path := ""
	
	if len(args) == 1 {
		path = args[0]
	} else if len(args) == 2 && args[0] == "-n" {
		if num, err := strconv.Atoi(args[1]); err == nil {
			lines = num
		}
	} else if len(args) == 3 && args[0] == "-n" {
		if num, err := strconv.Atoi(args[1]); err == nil {
			lines = num
			path = args[2]
		}
	}

	if path == "" {
		return "Usage: tail [-n] <file>"
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.TailFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path, lines)
	if err != nil {
		// Fallback to local processing
		fullContent, readErr := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
		if readErr != nil {
			return errorStyle.Render(fmt.Sprintf("Error reading file: %v", readErr))
		}
		content = ops.TailLocal(fullContent, lines)
	}

	return content
}

func (s *Shell) grepFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) < 2 {
		return "Usage: grep [-i] <pattern> <file>"
	}

	ignoreCase := false
	pattern := ""
	path := ""

	if len(args) == 2 {
		pattern, path = args[0], args[1]
	} else if len(args) == 3 && args[0] == "-i" {
		ignoreCase = true
		pattern, path = args[1], args[2]
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.GrepFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, pattern, path, ignoreCase)
	if err != nil {
		// Fallback to local processing
		fullContent, readErr := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
		if readErr != nil {
			return errorStyle.Render(fmt.Sprintf("Error reading file: %v", readErr))
		}
		content = ops.GrepLocal(fullContent, pattern, ignoreCase)
	}

	return content
}

func (s *Shell) wordCount(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) == 0 {
		return "Usage: wc <file>"
	}

	path := args[0]
	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.WordCount(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		// Fallback to local processing
		fullContent, readErr := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
		if readErr != nil {
			return errorStyle.Render(fmt.Sprintf("Error reading file: %v", readErr))
		}
		content = ops.WordCountLocal(fullContent)
	}

	return content
}

func (s *Shell) sortFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	unique := false
	path := ""

	if len(args) == 1 {
		path = args[0]
	} else if len(args) == 2 && args[0] == "-u" {
		unique = true
		path = args[1]
	}

	if path == "" {
		return "Usage: sort [-u] <file>"
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.SortFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path, unique)
	if err != nil {
		// Fallback to local processing
		fullContent, readErr := s.fileOps.ReadFile(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
		if readErr != nil {
			return errorStyle.Render(fmt.Sprintf("Error reading file: %v", readErr))
		}
		content = ops.SortLocal(fullContent, unique)
	}

	return content
}

// System information tools
func (s *Shell) listProcesses(ctx context.Context) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	content, err := s.fileOps.ListProcesses(ctx, s.currentNamespace, s.currentPod, s.currentContainer)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error listing processes: %v", err))
	}

	return content
}

func (s *Shell) getEnvironment(ctx context.Context) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	content, err := s.fileOps.GetEnvironment(ctx, s.currentNamespace, s.currentPod, s.currentContainer)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error getting environment: %v", err))
	}

	return content
}

func (s *Shell) getDiskUsage(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	path := "/"
	if len(args) > 0 {
		path = args[0]
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.GetDiskUsage(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error getting disk usage: %v", err))
	}

	return content
}

func (s *Shell) getDirectoryUsage(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	if !strings.HasPrefix(path, "/") {
		path = s.currentDir + "/" + path
	}

	content, err := s.fileOps.GetDirectoryUsage(ctx, s.currentNamespace, s.currentPod, s.currentContainer, path)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error getting directory usage: %v", err))
	}

	return content
}

func (s *Shell) getPodIP(ctx context.Context) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	pod, err := s.client.GetPod(ctx, s.currentNamespace, s.currentPod)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error getting pod info: %v", err))
	}

	if len(pod.Status.PodIPs) > 0 {
		var ips []string
		for _, podIP := range pod.Status.PodIPs {
			ips = append(ips, podIP.IP)
		}
		return fmt.Sprintf("Pod IPs: %s", strings.Join(ips, ", "))
	}

	if pod.Status.PodIP != "" {
		return fmt.Sprintf("Pod IP: %s", pod.Status.PodIP)
	}

	return "No IP address assigned to pod"
}

// Bubble Tea Model
func (m ShellModel) Init() tea.Cmd {
	return nil
}

func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmd := strings.TrimSpace(m.shell.input)
			if cmd != "" {
				ctx := context.Background()
				result := m.shell.ExecuteCommand(ctx, cmd)
				if result == "exit" {
					return m, tea.Quit
				}
				if result == "clear" {
					m.output = []string{}
				} else {
					m.output = append(m.output, fmt.Sprintf("$ %s", cmd))
					if result != "" {
						m.output = append(m.output, result)
					}
				}
				m.shell.input = ""
			}
		case tea.KeyBackspace:
			if len(m.shell.input) > 0 {
				m.shell.input = m.shell.input[:len(m.shell.input)-1]
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.shell.input += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m ShellModel) View() string {
	var output strings.Builder
	
	// Show output
	for _, line := range m.output {
		output.WriteString(line + "\n")
	}
	
	// Show prompt with context
	prompt := statusStyle.Render("k8sh")
	if m.shell.currentPod != "" {
		prompt += fmt.Sprintf("[%s/%s:%s]", m.shell.currentNamespace, m.shell.currentPod, m.shell.currentContainer)
	} else if m.shell.currentNamespace != "default" {
		prompt += fmt.Sprintf("[%s]", m.shell.currentNamespace)
	}
	prompt += " "
	prompt += promptStyle.Render(m.shell.currentDir) + " "
	prompt += m.shell.input
	
	return output.String()
}

func StartShell(client *k8s.Client) error {
	shell := NewShell(client)
	model := ShellModel{
		shell:  shell,
		output: []string{},
	}

	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
