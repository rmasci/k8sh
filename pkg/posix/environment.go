package posix

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Environment manages shell environment and state
type Environment struct {
	mu        sync.RWMutex
	vars      map[string]string
	exports   map[string]bool
	readonly  map[string]bool
	functions map[string]*Function
	aliases   map[string]string
	options   ShellOptions
	workingDir string
}

// ShellOptions contains shell configuration options
type ShellOptions struct {
	NoGlob       bool // -f: disable pathname expansion
	TrackAll     bool // -a: export all variables
	Notify       bool // -b: notify of job completion
	ErrExit      bool // -e: exit on error
	NoUnset      bool // -u: treat unset variables as error
	Verbose      bool // -v: verbose mode
	XTrace       bool // -x: trace commands
	NoExec       bool // -n: read commands but don't execute
	Interactive  bool // -i: interactive mode
	LoginShell   bool // -l: login shell
	Restricted  bool // -r: restricted shell
}

// Function represents a shell function
type Function struct {
	Name   string
	Body   []ASTNode
	Params []string
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	env := &Environment{
		vars:      make(map[string]string),
		exports:   make(map[string]bool),
		readonly:  make(map[string]bool),
		functions: make(map[string]*Function),
		aliases:   make(map[string]string),
		options:   ShellOptions{},
		workingDir: "/",
	}
	
	// Initialize with standard environment variables
	env.initializeStandardVars()
	
	return env
}

// initializeStandardVars sets up standard POSIX variables
func (env *Environment) initializeStandardVars() {
	// Copy current process environment
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			env.vars[parts[0]] = parts[1]
			env.exports[parts[0]] = true
		}
	}
	
	// Set shell-specific variables
	env.vars["0"] = "k8sh"
	env.vars["$"] = strconv.Itoa(os.Getpid())
	env.vars["PPID"] = strconv.Itoa(os.Getppid())
	env.vars["PWD"] = env.workingDir
	env.vars["SHLVL"] = "1"
	env.vars["IFS"] = " \t\n"
	
	// Export standard variables
	env.exports["PWD"] = true
	env.exports["SHLVL"] = true
	env.exports["IFS"] = true
}

// GetVar gets a variable value
func (env *Environment) GetVar(name string) (string, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	value, exists := env.vars[name]
	return value, exists
}

// SetVar sets a variable value
func (env *Environment) SetVar(name, value string, export, readonly bool) error {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	// Check if variable is readonly
	if env.readonly[name] {
		return fmt.Errorf("variable %s is readonly", name)
	}
	
	env.vars[name] = value
	if export {
		env.exports[name] = true
	}
	if readonly {
		env.readonly[name] = true
	}
	
	return nil
}

// ExportVar marks a variable for export
func (env *Environment) ExportVar(name string) error {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	if _, exists := env.vars[name]; !exists {
		if env.options.NoUnset {
			return fmt.Errorf("variable %s is not set", name)
		}
		env.vars[name] = ""
	}
	
	env.exports[name] = true
	return nil
}

// UnsetVar removes a variable
func (env *Environment) UnsetVar(name string) error {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	// Check if variable is readonly
	if env.readonly[name] {
		return fmt.Errorf("variable %s is readonly", name)
	}
	
	delete(env.vars, name)
	delete(env.exports, name)
	delete(env.readonly, name)
	
	return nil
}

// GetAlias gets an alias value
func (env *Environment) GetAlias(name string) (string, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	value, exists := env.aliases[name]
	return value, exists
}

// SetAlias sets an alias
func (env *Environment) SetAlias(name, value string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	env.aliases[name] = value
}

// UnsetAlias removes an alias
func (env *Environment) UnsetAlias(name string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	delete(env.aliases, name)
}

// GetFunction gets a function definition
func (env *Environment) GetFunction(name string) (*Function, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	funcDef, exists := env.functions[name]
	return funcDef, exists
}

// SetFunction sets a function definition
func (env *Environment) SetFunction(name string, body []ASTNode, params []string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	env.functions[name] = &Function{
		Name:   name,
		Body:   body,
		Params: params,
	}
}

// GetWorkingDir returns the current working directory
func (env *Environment) GetWorkingDir() string {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	return env.workingDir
}

// SetWorkingDir sets the current working directory
func (env *Environment) SetWorkingDir(dir string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	env.workingDir = dir
	env.vars["PWD"] = dir
}

// GetOptions returns current shell options
func (env *Environment) GetOptions() ShellOptions {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	return env.options
}

// SetOption sets a shell option
func (env *Environment) SetOption(option string, value bool) {
	env.mu.Lock()
	defer env.mu.Unlock()
	
	switch option {
	case "f":
		env.options.NoGlob = value
	case "a":
		env.options.TrackAll = value
	case "b":
		env.options.Notify = value
	case "e":
		env.options.ErrExit = value
	case "u":
		env.options.NoUnset = value
	case "v":
		env.options.Verbose = value
	case "x":
		env.options.XTrace = value
	case "n":
		env.options.NoExec = value
	case "i":
		env.options.Interactive = value
	case "l":
		env.options.LoginShell = value
	case "r":
		env.options.Restricted = value
	}
}

// ExpandVariables performs variable expansion in a string
func (env *Environment) ExpandVariables(input string) string {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	result := input
	i := 0
	
	for i < len(result) {
		if result[i] == '$' && i+1 < len(result) {
			// Check for ${var} syntax
			if result[i+1] == '{' {
				end := strings.Index(result[i:], "}")
				if end != -1 {
					varName := result[i+2 : i+end]
					if value, exists := env.vars[varName]; exists {
						result = result[:i] + value + result[i+end+1:]
						i += len(value)
					} else {
						if env.options.NoUnset {
							result = result[:i] + result[i+end+1:]
						} else {
							i += end + 1
						}
					}
					continue
				}
			}
			
			// Check for $var syntax
			varName := ""
			j := i + 1
			for j < len(result) && (isAlphaNum(result[j]) || result[j] == '_') {
				varName += string(result[j])
				j++
			}
			
			if varName != "" {
				if value, exists := env.vars[varName]; exists {
					result = result[:i] + value + result[j:]
					i += len(value)
				} else {
					if env.options.NoUnset {
						result = result[:i] + result[j:]
					} else {
						i = j
					}
				}
				continue
			}
			
			// Check for special parameters
			if j < len(result) {
				switch result[j] {
				case '$':
					result = result[:i] + strconv.Itoa(os.Getpid()) + result[j+1:]
					i += len(strconv.Itoa(os.Getpid()))
				case '?':
					// TODO: Track exit status
					result = result[:i] + "0" + result[j+1:]
					i++
				case '!':
					// TODO: Track background job PIDs
					result = result[:i] + "" + result[j+1:]
				case '#':
					// TODO: Track argument count
					result = result[:i] + "0" + result[j+1:]
					i++
				default:
					i++
				}
				continue
			}
		}
		i++
	}
	
	return result
}

// GetExportedVars returns all exported variables
func (env *Environment) GetExportedVars() map[string]string {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	exported := make(map[string]string)
	for name, isExported := range env.exports {
		if isExported {
			if value, exists := env.vars[name]; exists {
				exported[name] = value
			}
		}
	}
	
	return exported
}

// Clone creates a copy of the environment
func (env *Environment) Clone() *Environment {
	env.mu.RLock()
	defer env.mu.RUnlock()
	
	clone := &Environment{
		vars:      make(map[string]string),
		exports:   make(map[string]bool),
		readonly:  make(map[string]bool),
		functions: make(map[string]*Function),
		aliases:   make(map[string]string),
		options:   env.options,
		workingDir: env.workingDir,
	}
	
	// Copy variables
	for k, v := range env.vars {
		clone.vars[k] = v
	}
	for k, v := range env.exports {
		clone.exports[k] = v
	}
	for k, v := range env.readonly {
		clone.readonly[k] = v
	}
	
	// Copy functions
	for k, v := range env.functions {
		clone.functions[k] = &Function{
			Name:   v.Name,
			Body:   v.Body,
			Params: v.Params,
		}
	}
	
	// Copy aliases
	for k, v := range env.aliases {
		clone.aliases[k] = v
	}
	
	return clone
}

// Helper functions
func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
