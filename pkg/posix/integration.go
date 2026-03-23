package posix

import (
	"context"
	"fmt"
	"strings"

	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/shell"
)

// POSIXShell extends the existing k8sh shell with POSIX compliance
type POSIXShell struct {
	*shell.Shell
	posixEnv *Environment
	executor *Executor
}

// NewPOSIXShell creates a new POSIX-compliant shell
func NewPOSIXShell(client *k8s.Client) *POSIXShell {
	k8shShell := shell.NewShell(client)
	posixEnv := NewEnvironment()
	executor := NewExecutor(posixEnv)
	
	// Set up integration between POSIX and k8sh
	posixShell := &POSIXShell{
		Shell:    k8shShell,
		posixEnv: posixEnv,
		executor: executor,
	}
	
	// Initialize POSIX environment with k8sh context
	posixShell.initializeIntegration()
	
	return posixShell
}

// initializeIntegration sets up the integration between POSIX and k8sh
func (ps *POSIXShell) initializeIntegration() {
	// Sync working directory
	ps.posixEnv.SetWorkingDir(ps.GetCurrentDir())
	
	// Set up k8sh-specific variables
	ps.posixEnv.SetVar("K8SH_POD", ps.GetCurrentPod(), false, false)
	ps.posixEnv.SetVar("K8SH_CONTAINER", ps.GetCurrentContainer(), false, false)
	ps.posixEnv.SetVar("K8SH_NAMESPACE", ps.GetCurrentNamespace(), false, false)
	
	// Export k8sh variables
	ps.posixEnv.ExportVar("K8SH_POD")
	ps.posixEnv.ExportVar("K8SH_CONTAINER")
	ps.posixEnv.ExportVar("K8SH_NAMESPACE")
}

// ExecuteCommand executes a command with POSIX parsing
func (ps *POSIXShell) ExecuteCommand(ctx context.Context, cmd string) string {
	// Try POSIX parsing first
	parser := NewParser(cmd)
	ast, err := parser.Parse()
	
	if err != nil {
		// Fallback to k8sh parsing if POSIX parsing fails
		return ps.Shell.ExecuteCommand(ctx, cmd)
	}
	
	// Execute with POSIX executor
	result, err := ps.executor.Execute(ctx, ast)
	if err != nil {
		// Check if it's a special control flow error
		if ps.isControlFlowError(err) {
			return ps.handleControlFlow(ctx, err)
		}
		
		// Fallback to k8sh for unknown commands
		if ps.isUnknownCommandError(err) {
			return ps.Shell.ExecuteCommand(ctx, cmd)
		}
		
		return fmt.Sprintf("Error: %v\n", err)
	}
	
	// Sync state back to k8sh
	ps.syncState()
	
	return result
}

// isControlFlowError checks if an error represents control flow
func (ps *POSIXShell) isControlFlowError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "exit") || 
		   strings.Contains(errStr, "return") ||
		   strings.Contains(errStr, "break") ||
		   strings.Contains(errStr, "continue")
}

// isUnknownCommandError checks if an error indicates an unknown command
func (ps *POSIXShell) isUnknownCommandError(err error) bool {
	return strings.Contains(err.Error(), "not found") ||
		   strings.Contains(err.Error(), "not implemented")
}

// handleControlFlow handles special control flow commands
func (ps *POSIXShell) handleControlFlow(ctx context.Context, err error) string {
	errStr := err.Error()
	
	if strings.Contains(errStr, "exit") {
		return "exit"
	}
	
	// For other control flow, return the error message
	return errStr + "\n"
}

// syncState synchronizes POSIX state back to k8sh
func (ps *POSIXShell) syncState() {
	// Sync working directory
	posixWD := ps.posixEnv.GetWorkingDir()
	if posixWD != ps.GetCurrentDir() {
		// Update k8sh working directory
		ps.SetCurrentDir(posixWD)
	}
	
	// Sync environment variables that affect k8sh
	if pod, exists := ps.posixEnv.GetVar("K8SH_POD"); exists {
		if pod != ps.GetCurrentPod() {
			ps.SetCurrentPod(pod)
		}
	}
	
	if namespace, exists := ps.posixEnv.GetVar("K8SH_NAMESPACE"); exists {
		if namespace != ps.GetCurrentNamespace() {
			ps.SetCurrentNamespace(namespace)
		}
	}
}

// Extended POSIX-specific methods

// ExecuteScript executes a shell script
func (ps *POSIXShell) ExecuteScript(ctx context.Context, script string) (string, error) {
	lines := strings.Split(script, "\n")
	var output strings.Builder
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Execute line
		result := ps.ExecuteCommand(ctx, line)
		
		// Check for exit
		if result == "exit" {
			break
		}
		
		// Collect output
		if result != "" {
			output.WriteString(result)
			if !strings.HasSuffix(result, "\n") {
				output.WriteString("\n")
			}
		}
		
		// Check for error exit if errexit is set
		if ps.posixEnv.GetOptions().ErrExit && strings.Contains(result, "Error:") {
			return output.String(), fmt.Errorf("script failed at line %d: %s", i+1, line)
		}
	}
	
	return output.String(), nil
}

// SetPOSIXOption sets a POSIX shell option
func (ps *POSIXShell) SetPOSIXOption(option string, value bool) {
	ps.posixEnv.SetOption(option, value)
}

// GetPOSIXOptions returns current POSIX options
func (ps *POSIXShell) GetPOSIXOptions() ShellOptions {
	return ps.posixEnv.GetOptions()
}

// SetEnvironmentVariable sets an environment variable
func (ps *POSIXShell) SetEnvironmentVariable(name, value string, export bool) error {
	return ps.posixEnv.SetVar(name, value, export, false)
}

// GetEnvironmentVariable gets an environment variable
func (ps *POSIXShell) GetEnvironmentVariable(name string) (string, bool) {
	return ps.posixEnv.GetVar(name)
}

// SetAlias sets a command alias
func (ps *POSIXShell) SetAlias(name, value string) {
	ps.posixEnv.SetAlias(name, value)
}

// GetAlias gets an alias value
func (ps *POSIXShell) GetAlias(name string) (string, bool) {
	return ps.posixEnv.GetAlias(name)
}

// DefineFunction defines a shell function
func (ps *POSIXShell) DefineFunction(name string, body string, params []string) error {
	parser := NewParser(body)
	ast, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("function parse error: %v", err)
	}
	
	// Convert AST to function body
	var nodes []ASTNode
	if cmdNode, ok := ast.(*CommandNode); ok {
		nodes = []ASTNode{cmdNode}
	} else {
		nodes = []ASTNode{ast}
	}
	
	ps.posixEnv.SetFunction(name, nodes, params)
	return nil
}

// GetJobs returns background jobs
func (ps *POSIXShell) GetJobs() map[int]*Job {
	return ps.executor.GetJobs()
}

// GetWorkingDir returns the current working directory
func (ps *POSIXShell) GetWorkingDir() string {
	return ps.posixEnv.GetWorkingDir()
}

// SetWorkingDir sets the current working directory
func (ps *POSIXShell) SetWorkingDir(dir string) {
	ps.posixEnv.SetWorkingDir(dir)
	ps.SetCurrentDir(dir)
}

// K8sh integration methods

// GetCurrentDir returns the current directory from k8sh
func (ps *POSIXShell) GetCurrentDir() string {
	// This would need to be implemented in the k8sh shell
	return "/"
}

// SetCurrentDir sets the current directory in k8sh
func (ps *POSIXShell) SetCurrentDir(dir string) {
	// This would need to be implemented in the k8sh shell
}

// GetCurrentPod returns the current pod from k8sh
func (ps *POSIXShell) GetCurrentPod() string {
	// This would need to be implemented in the k8sh shell
	return ""
}

// SetCurrentPod sets the current pod in k8sh
func (ps *POSIXShell) SetCurrentPod(pod string) {
	// This would need to be implemented in the k8sh shell
}

// GetCurrentContainer returns the current container from k8sh
func (ps *POSIXShell) GetCurrentContainer() string {
	// This would need to be implemented in the k8sh shell
	return ""
}

// SetCurrentContainer sets the current container in k8sh
func (ps *POSIXShell) SetCurrentContainer(container string) {
	// This would need to be implemented in the k8sh shell
}

// GetCurrentNamespace returns the current namespace from k8sh
func (ps *POSIXShell) GetCurrentNamespace() string {
	// This would need to be implemented in the k8sh shell
	return "default"
}

// SetCurrentNamespace sets the current namespace in k8sh
func (ps *POSIXShell) SetCurrentNamespace(namespace string) {
	// This would need to be implemented in the k8sh shell
}
