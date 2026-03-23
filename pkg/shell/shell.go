package shell

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/ops"
	"github.com/rmasci/k8sh/pkg/editor"
	"k8s.io/client-go/tools/clientcmd"
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
	historyIndex  int
	input         string
	suggestions   []string
	suggestionIndex int
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
		historyIndex:  -1,
		input:         "",
		suggestions:   []string{},
		suggestionIndex: -1,
	}
}

func (s *Shell) addToHistory(cmd string) {
	// Don't add empty commands or duplicates
	if cmd == "" {
		return
	}
	
	// Remove from history if it already exists (to avoid duplicates)
	for i, h := range s.history {
		if h == cmd {
			s.history = append(s.history[:i], s.history[i+1:]...)
			break
		}
	}
	
	// Add to history
	s.history = append(s.history, cmd)
	
	// Limit history size
	if len(s.history) > 100 {
		s.history = s.history[1:]
	}
	
	// Reset history index
	s.historyIndex = -1
}

func (s *Shell) getCompletions(input string) []string {
	var completions []string
	
	// Get command suggestions
	commands := []string{
		"help", "exit", "quit", "pwd", "ls", "cd", "cat", "vi", "vim",
		"mkdir", "rm", "cp", "download", "mv", "touch", "head", "tail",
		"grep", "wc", "sort", "ps", "env", "df", "du", "ip", "pods",
		"use", "namespace", "clear",
	}
	
	// If input is empty, show all commands
	if input == "" {
		return commands
	}
	
	// Complete commands
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			completions = append(completions, cmd)
		}
	}
	
	// Complete pod names if using "use" command
	if strings.HasPrefix(input, "use ") {
		podPart := strings.TrimPrefix(input, "use ")
		// Get pod completions
		ctx := context.Background()
		pods, err := s.client.ListPods(ctx, s.currentNamespace)
		if err == nil {
			for _, pod := range pods {
				if strings.HasPrefix(pod.Name, podPart) {
					completions = append(completions, "use "+pod.Name)
				}
			}
		}
	}
	
	return completions
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
	case "download":
		return s.downloadFile(ctx, args...)
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
	return `🐚 k8sh - Kubernetes Pseudo-Shell
=====================================

k8sh provides a shell interface for Kubernetes pods without requiring
any tools to be installed in the target containers. Perfect for 
distroless, scratch, and minimal container environments!

🚀 QUICK START:
  1. List available pods:     k8sh> pods
  2. Select a pod:           k8sh> use my-pod
  3. Start working:           k8sh> ls -la
  4. Get help anytime:        k8sh> help

⌨️  SHELL FEATURES:
  • Tab Completion: Press Tab to complete commands and pod names
  • History Navigation: Use ↑/↓ arrows to browse command history
  • Smart Suggestions: Context-aware command suggestions
  • Duplicate Prevention: Intelligent history management
  • Visual Feedback: Highlighted suggestions and current selection

⚙️  CONFIGURATION:
  k8sh uses standard Kubernetes configuration at:
    ~/.kube/config
  
  First-time setup options:
  • Local cluster: minikube start, kind create cluster
  • Cloud provider: gcloud init, aws eks update-kubeconfig
  • Manual setup: kubectl config commands
  
  Current config check: k8sh> pods (shows config status)

📁 FILE OPERATIONS:
  mkdir <path>           Create directory
    Example: mkdir /app/data
  
  rm [-r] <path>         Remove file/directory
    Example: rm -r /tmp/old-data
    Example: rm /tmp/single-file.txt
  
  cp <src> <dst>         Copy file within pod
    Example: cp config.yaml config.yaml.bak
  
  download <src> <dst>    🆕 Download file from pod to local
    Example: download /app/logs/app.log ./logs-backup.log
    Example: download /app/config/ ./config-backup/ -r
  
  mv <src> <dst>         Move/rename file
    Example: mv old-name.txt new-name.txt
  
  touch <file>           Create empty file
    Example: touch /tmp/new-file.txt

📖 TEXT EDITING:
  cat <file>             Display file contents
    Example: cat /app/config.yaml
  
  vi/vim <file>          Edit file with vi editor
    Example: vi /app/config.yaml
    Navigation: h,j,k,l, w,b, 0,$, gg/G
    Editing: i/a/o/O, x/dd/dw
    Search: /pattern, n/N for next/prev
    Save/Quit: :w, :q, :wq

📊 TEXT PROCESSING:
  head [-n] <file>       Show first lines
    Example: head -20 /app/logs/app.log
  
  tail [-n] <file>       Show last lines  
    Example: tail -50 /app/logs/app.log
  
  grep [-i] <pattern> <file> Search in files
    Example: grep -i "error" /app/logs/*.log
  
  wc <file>              Word/line/character count
    Example: wc /app/logs/app.log
  
  sort [-u] <file>       Sort lines (unique with -u)
    Example: sort -u /app/users.txt

🔍 SYSTEM INFORMATION:
  ps                      List running processes
    Example: ps aux
  
  env                     Show environment variables
    Example: env | grep APP_
  
  df [path]              Disk usage by filesystem
    Example: df /app
  
  du [path]              Directory usage
    Example: du -sh /app/logs
  
  ip                      Show pod IP address
    Example: ip

☸️  KUBERNETES COMMANDS:
  pods                   List all available pods
    Example: pods
    Example: pods --namespace=production
  
  use <pod> [container]  Select pod and container
    Example: use my-app
    Example: use my-app container-name
  
  namespace <name>       Set/change namespace
    Example: namespace production
    Example: namespace

🔄 OTHER:
  clear                  Clear terminal screen
  help                  Show this help message
  exit/quit             Exit the shell

💡 PRO TIPS:
  • Use 'download' to copy files from pod to your local machine
  • All file paths work with both absolute (/path) and relative (path) formats
  • Tab completion works for pod and container names
  • Use 'pods' command to see all available containers
  • The vi editor supports full vi keybindings and commands

🐚 POSIX MODE:
  For full POSIX compliance (pipelines, redirection, variables):
    k8sh posix
  
  POSIX features include:
  • Command pipelines: cmd1 | cmd2 | cmd3
  • I/O redirection: >, >>, <, 2>
  • Variable expansion: VAR, ${VAR}
  • Command substitution: $(cmd)
  • Built-in commands: echo, printf, export, cd, pwd, etc.

📚 FOR MORE HELP:
  • GitHub: https://github.com/rmasci/k8sh
  • Issues: https://github.com/rmasci/k8sh/issues
  • POSIX Guide: See docs/posix_compliance.md

Happy container hacking! 🎉
`
}

func (s *Shell) listPods(ctx context.Context) string {
	if s.currentPod == "" {
		pods, err := s.client.ListPods(ctx, s.currentNamespace)
		if err != nil {
			// Check if this is a config error
			if strings.Contains(err.Error(), "kube/config") || strings.Contains(err.Error(), "no such file") {
				return errorStyle.Render(fmt.Sprintf(`🔍 KUBERNETES CONFIG NOT FOUND

k8sh looks for Kubernetes configuration at:
  %s

🚀 FIRST-TIME SETUP:
==================

1. 📁 LOCAL CLUSTER (minikube, kind, etc.):
   minikube start
   kind create cluster

2. ☁️  CLOUD PROVIDER:
   gcloud init
   aws eks update-kubeconfig  
   az account set

3. 🔧 MANUAL CONFIG:
   mkdir -p %s
   # Edit the config file with your cluster details

4. 📋 IN-CLUSTER CONFIG:
   kubectl config set-cluster my-cluster --server=https://...
   kubectl config set-credentials my-user --token=...
   kubectl config set-context my-context --cluster=my-cluster --user=my-user
   kubectl config use-context my-context

After setup, run k8sh again! 🎉`, clientcmd.RecommendedHomeFile, filepath.Dir(clientcmd.RecommendedHomeFile)))
			}
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

func (s *Shell) downloadFile(ctx context.Context, args ...string) string {
	if s.currentPod == "" {
		return errorStyle.Render("No pod selected. Use 'use <pod>' to select a pod first")
	}

	if len(args) < 2 {
		return "Usage: download <src> <dst> [-r|--recursive]"
	}

	// Parse arguments
	var recursive bool
	var srcPath, dstPath string
	remainingArgs := []string{}
	
	for i, arg := range args {
		if arg == "-r" || arg == "--recursive" {
			recursive = true
		} else if i == 0 {
			srcPath = arg
		} else if i == 1 {
			dstPath = arg
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// Handle relative paths
	if !strings.HasPrefix(srcPath, "/") {
		srcPath = s.currentDir + "/" + srcPath
	}

	// Use k8s client to download file directly
	err := s.client.DownloadFile(s.currentNamespace, s.currentPod, s.currentContainer, srcPath, dstPath, recursive)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error downloading file: %v", err))
	}

	return fmt.Sprintf("Downloaded: %s -> %s", srcPath, dstPath)
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

// Getter methods for POSIX integration
func (s *Shell) GetCurrentDir() string {
	return s.currentDir
}

func (s *Shell) GetCurrentPod() string {
	return s.currentPod
}

func (s *Shell) GetCurrentContainer() string {
	return s.currentContainer
}

func (s *Shell) GetCurrentNamespace() string {
	return s.currentNamespace
}

func (s *Shell) SetCurrentDir(dir string) {
	s.currentDir = dir
}

func (s *Shell) SetCurrentPod(pod string) {
	s.currentPod = pod
}

func (s *Shell) SetCurrentContainer(container string) {
	s.currentContainer = container
}

func (s *Shell) SetCurrentNamespace(namespace string) {
	s.currentNamespace = namespace
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
				// Add to history
				m.shell.addToHistory(cmd)
				
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
				m.shell.suggestions = []string{}
				m.shell.suggestionIndex = -1
			}
		case tea.KeyTab:
			// Tab completion
			if len(m.shell.suggestions) == 0 {
				// Generate suggestions
				m.shell.suggestions = m.shell.getCompletions(m.shell.input)
				m.shell.suggestionIndex = 0
			} else {
				// Cycle through suggestions
				m.shell.suggestionIndex++
				if m.shell.suggestionIndex >= len(m.shell.suggestions) {
					m.shell.suggestionIndex = 0
				}
			}
			// Apply suggestion
			if len(m.shell.suggestions) > 0 {
				m.shell.input = m.shell.suggestions[m.shell.suggestionIndex]
			}
		case tea.KeyUp:
			// Navigate history up
			if len(m.shell.history) > 0 {
				if m.shell.historyIndex == -1 {
					m.shell.historyIndex = len(m.shell.history) - 1
				} else if m.shell.historyIndex > 0 {
					m.shell.historyIndex--
				}
				m.shell.input = m.shell.history[m.shell.historyIndex]
				m.shell.suggestions = []string{}
				m.shell.suggestionIndex = -1
			}
		case tea.KeyDown:
			// Navigate history down
			if m.shell.historyIndex != -1 {
				m.shell.historyIndex++
				if m.shell.historyIndex >= len(m.shell.history) {
					m.shell.historyIndex = -1
					m.shell.input = ""
				} else {
					m.shell.input = m.shell.history[m.shell.historyIndex]
				}
				m.shell.suggestions = []string{}
				m.shell.suggestionIndex = -1
			}
		case tea.KeyBackspace:
			if len(m.shell.input) > 0 {
				m.shell.input = m.shell.input[:len(m.shell.input)-1]
				// Clear suggestions when input changes
				m.shell.suggestions = []string{}
				m.shell.suggestionIndex = -1
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.shell.input += string(msg.Runes)
				// Clear suggestions when input changes
				m.shell.suggestions = []string{}
				m.shell.suggestionIndex = -1
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
	
	// Show suggestions if available
	if len(m.shell.suggestions) > 0 {
		output.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Suggestions:"))
		for i, suggestion := range m.shell.suggestions {
			if i == m.shell.suggestionIndex {
				output.WriteString("\n  → " + lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render(suggestion))
			} else {
				output.WriteString("\n    " + suggestion)
			}
		}
		output.WriteString("\n")
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
	
	// Show cursor indicator
	if len(m.shell.suggestions) > 0 {
		prompt += " (Tab for more)"
	}
	
	output.WriteString(prompt)
	
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
