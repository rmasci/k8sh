package posix

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Executor handles execution of parsed AST nodes
type Executor struct {
	env      *Environment
	builtins *Builtins
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	jobs     map[int]*Job
	nextJob  int
	mu       sync.Mutex
}

// Job represents a background job
type Job struct {
	ID       int
	Command  string
	PID      int
	Status   string
	Started  bool
	Finished bool
}

// NewExecutor creates a new executor
func NewExecutor(env *Environment) *Executor {
	return &Executor{
		env:      env,
		builtins: NewBuiltins(),
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		jobs:     make(map[int]*Job),
		nextJob:  1,
	}
}

// Execute executes an AST node
func (e *Executor) Execute(ctx context.Context, node ASTNode) (string, error) {
	switch n := node.(type) {
	case *StringNode:
		return n.Value, nil
	case *VariableNode:
		if value, exists := e.env.GetVar(n.Name); exists {
			return value, nil
		}
		if e.env.GetOptions().NoUnset {
			return "", fmt.Errorf("variable %s is not set", n.Name)
		}
		return "", nil
	case *SubstitutionNode:
		return e.executeSubstitution(ctx, n)
	case *CommandNode:
		return e.executeCommand(ctx, n)
	case *PipelineNode:
		return e.executePipeline(ctx, n)
	case *IfNode:
		return e.executeIf(ctx, n)
	case *ForNode:
		return e.executeFor(ctx, n)
	case *WhileNode:
		return e.executeWhile(ctx, n)
	default:
		return "", fmt.Errorf("unknown node type: %T", n)
	}
}

// executeCommand executes a single command
func (e *Executor) executeCommand(ctx context.Context, cmd *CommandNode) (string, error) {
	// Expand variables in command and arguments
	expandedArgs := []string{}
	
	// Expand command name
	expandedCmd := e.env.ExpandVariables(cmd.Command)
	
	// Check for aliases
	if alias, exists := e.env.GetAlias(expandedCmd); exists {
		// Parse and execute the aliased command
		parser := NewParser(alias)
		ast, err := parser.Parse()
		if err != nil {
			return "", fmt.Errorf("alias parse error: %v", err)
		}
		return e.Execute(ctx, ast)
	}
	
	// Expand arguments
	for _, arg := range cmd.Args {
		var expanded string
		switch node := arg.(type) {
		case *StringNode:
			expanded = node.Value
		case *VariableNode:
			if value, exists := e.env.GetVar(node.Name); exists {
				expanded = value
			} else {
				expanded = ""
			}
		case *SubstitutionNode:
			result, err := e.Execute(ctx, node.Command)
			if err != nil {
				return "", err
			}
			expanded = strings.TrimSuffix(result, "\n")
		default:
			expanded = arg.String()
		}
		expanded = e.env.ExpandVariables(expanded)
		expandedArgs = append(expandedArgs, expanded)
	}
	
	// Check for builtin functions
	if builtin, exists := e.builtins.GetBuiltin(expandedCmd); exists {
		return builtin(ctx, e.env, expandedArgs)
	}
	
	// Check for functions
	if fn, exists := e.env.GetFunction(expandedCmd); exists {
		return e.executeFunction(ctx, fn, expandedArgs)
	}
	
	// Execute external command (delegated to k8sh shell)
	return e.executeExternalCommand(ctx, expandedCmd, expandedArgs, cmd.Redirs, cmd.Background)
}

// executePipeline executes a pipeline of commands
func (e *Executor) executePipeline(ctx context.Context, pipeline *PipelineNode) (string, error) {
	if len(pipeline.Commands) == 1 {
		return e.Execute(ctx, pipeline.Commands[0])
	}
	
	// Create pipes between commands
	pipes := make([]io.Reader, len(pipeline.Commands)-1)
	for i := range pipes {
		pr, pw := io.Pipe()
		pipes[i] = pr
		defer pw.Close()
		defer pr.Close()
	}
	
	var results []string
	var lastError error
	
	// Execute each command in the pipeline
	for i, cmdNode := range pipeline.Commands {
		cmd, ok := cmdNode.(*CommandNode)
		if !ok {
			return "", fmt.Errorf("pipeline contains non-command: %T", cmdNode)
		}
		
		// Create a copy of the executor for this command
		cmdExecutor := e.clone()
		
		// Set up stdin/stdout for piping
		if i > 0 {
			cmdExecutor.stdin = pipes[i-1]
		}
		if i < len(pipeline.Commands)-1 {
			// This is a simplification - in reality, we'd need to handle pipe writing
			cmdExecutor.stdout = io.Discard
		}
		
		// Execute the command
		result, err := cmdExecutor.Execute(ctx, cmd)
		results = append(results, result)
		if err != nil && lastError == nil {
			lastError = err
		}
	}
	
	// Return the output of the last command in the pipeline
	if len(results) > 0 {
		return results[len(results)-1], lastError
	}
	
	return "", lastError
}

// executeSubstitution executes command substitution
func (e *Executor) executeSubstitution(ctx context.Context, sub *SubstitutionNode) (string, error) {
	result, err := e.Execute(ctx, sub.Command)
	if err != nil {
		return "", err
	}
	
	// Trim trailing newline
	return strings.TrimSuffix(result, "\n"), nil
}

// executeIf executes an if statement
func (e *Executor) executeIf(ctx context.Context, ifNode *IfNode) (string, error) {
	// Execute condition
	_, err := e.Execute(ctx, ifNode.Condition)
	conditionSuccess := err == nil
	
	if conditionSuccess {
		// Execute then block
		for _, stmt := range ifNode.ThenBlock {
			result, err := e.Execute(ctx, stmt)
			if err != nil {
				return result, err
			}
		}
	} else {
		// Execute else block
		for _, stmt := range ifNode.ElseBlock {
			result, err := e.Execute(ctx, stmt)
			if err != nil {
				return result, err
			}
		}
	}
	
	return "", nil
}

// executeFor executes a for loop
func (e *Executor) executeFor(ctx context.Context, forNode *ForNode) (string, error) {
	var values []string
	
	// Get values to iterate over
	for _, valueNode := range forNode.Values {
		value, err := e.Execute(ctx, valueNode)
		if err != nil {
			return "", err
		}
		values = append(values, strings.Fields(value)...)
	}
	
	// Execute loop body for each value
	for _, value := range values {
		// Set loop variable
		e.env.SetVar(forNode.Variable, value, false, false)
		
		// Execute body
		for _, stmt := range forNode.Body {
			result, err := e.Execute(ctx, stmt)
			if err != nil {
				// Check if it's a break or continue
				if strings.Contains(err.Error(), "break") {
					return "", nil
				}
				if strings.Contains(err.Error(), "continue") {
					break
				}
				return result, err
			}
		}
	}
	
	return "", nil
}

// executeWhile executes a while loop
func (e *Executor) executeWhile(ctx context.Context, whileNode *WhileNode) (string, error) {
	for {
		// Execute condition
		_, err := e.Execute(ctx, whileNode.Condition)
		conditionSuccess := err == nil
		
		if !conditionSuccess {
			break
		}
		
		// Execute body
		for _, stmt := range whileNode.Body {
			result, err := e.Execute(ctx, stmt)
			if err != nil {
				// Check if it's a break or continue
				if strings.Contains(err.Error(), "break") {
					return "", nil
				}
				if strings.Contains(err.Error(), "continue") {
					break
				}
				return result, err
			}
		}
	}
	
	return "", nil
}

// executeFunction executes a user-defined function
func (e *Executor) executeFunction(ctx context.Context, fn *Function, args []string) (string, error) {
	// Create a new environment for the function
	funcEnv := e.env.Clone()
	
	// Set positional parameters
	funcEnv.SetVar("0", fn.Name, false, false)
	for i, arg := range args {
		funcEnv.SetVar(fmt.Sprintf("%d", i+1), arg, false, false)
	}
	
	// Create a new executor for the function
	funcExecutor := NewExecutor(funcEnv)
	funcExecutor.builtins = e.builtins
	funcExecutor.stdin = e.stdin
	funcExecutor.stdout = e.stdout
	funcExecutor.stderr = e.stderr
	
	// Execute function body
	var result string
	var err error
	
	for _, stmt := range fn.Body {
		result, err = funcExecutor.Execute(ctx, stmt)
		if err != nil {
			if strings.Contains(err.Error(), "return") {
				// Extract return code
				parts := strings.Fields(err.Error())
				if len(parts) > 1 {
					if code, parseErr := strconv.Atoi(parts[1]); parseErr == nil {
						e.env.SetVar("?", fmt.Sprintf("%d", code), false, false)
					}
				}
				return result, nil
			}
			return result, err
		}
	}
	
	return result, nil
}

// executeExternalCommand delegates to the k8sh shell for external commands
func (e *Executor) executeExternalCommand(ctx context.Context, cmd string, args []string, redirs []Redirection, background bool) (string, error) {
	// This is where we integrate with the existing k8sh shell
	// For now, return a placeholder implementation
	
	// Handle redirections
	for _, redir := range redirs {
		switch redir.Type {
		case RedirOutput, RedirAppend, RedirInput, RedirError, RedirAll:
			// TODO: Implement proper redirection handling
		}
	}
	
	if background {
		// Create background job
		job := &Job{
			ID:      e.nextJob,
			Command: fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")),
			Status:  "Running",
			Started: true,
		}
		
		e.mu.Lock()
		e.jobs[e.nextJob] = job
		e.nextJob++
		e.mu.Unlock()
		
		return fmt.Sprintf("[%d] %d\n", job.ID, job.PID), nil
	}
	
	// For now, return a placeholder
	return fmt.Sprintf("External command: %s %s\n", cmd, strings.Join(args, " ")), nil
}

// clone creates a copy of the executor
func (e *Executor) clone() *Executor {
	clone := &Executor{
		env:      e.env.Clone(),
		builtins: e.builtins,
		stdin:    e.stdin,
		stdout:   e.stdout,
		stderr:   e.stderr,
		jobs:     make(map[int]*Job),
		nextJob:  e.nextJob,
	}
	
	e.mu.Lock()
	for id, job := range e.jobs {
		clone.jobs[id] = &Job{
			ID:       job.ID,
			Command:  job.Command,
			PID:      job.PID,
			Status:   job.Status,
			Started:  job.Started,
			Finished: job.Finished,
		}
	}
	e.mu.Unlock()
	
	return clone
}

// GetJobs returns all background jobs
func (e *Executor) GetJobs() map[int]*Job {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	jobs := make(map[int]*Job)
	for id, job := range e.jobs {
		jobs[id] = &Job{
			ID:       job.ID,
			Command:  job.Command,
			PID:      job.PID,
			Status:   job.Status,
			Started:  job.Started,
			Finished: job.Finished,
		}
	}
	
	return jobs
}

// SetIO sets the I/O streams
func (e *Executor) SetIO(stdin io.Reader, stdout, stderr io.Writer) {
	e.stdin = stdin
	e.stdout = stdout
	e.stderr = stderr
}
