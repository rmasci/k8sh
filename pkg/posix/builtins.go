package posix

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BuiltinFunc represents a builtin function
type BuiltinFunc func(ctx context.Context, env *Environment, args []string) (string, error)

// Builtins contains all POSIX builtin functions
type Builtins struct {
	functions map[string]BuiltinFunc
}

// NewBuiltins creates a new builtin registry
func NewBuiltins() *Builtins {
	b := &Builtins{
		functions: make(map[string]BuiltinFunc),
	}
	
	// Register all builtin functions
	b.registerBuiltins()
	
	return b
}

// registerBuiltins registers all POSIX builtin functions
func (b *Builtins) registerBuiltins() {
	// Core builtins
	b.functions["echo"] = b.builtinEcho
	b.functions["printf"] = b.builtinPrintf
	b.functions["export"] = b.builtinExport
	b.functions["unset"] = b.builtinUnset
	b.functions["readonly"] = b.builtinReadonly
	b.functions["set"] = b.builtinSet
	b.functions["return"] = b.builtinReturn
	b.functions["exit"] = b.builtinExit
	b.functions["break"] = b.builtinBreak
	b.functions["continue"] = b.builtinContinue
	b.functions["colon"] = b.builtinColon // :
	b.functions["true"] = b.builtinTrue
	b.functions["false"] = b.builtinFalse
	
	// Alias and function builtins
	b.functions["alias"] = b.builtinAlias
	b.functions["unalias"] = b.builtinUnalias
	b.functions["type"] = b.builtinType
	b.functions["command"] = b.builtinCommand
	
	// Directory builtins
	b.functions["cd"] = b.builtinCD
	b.functions["pwd"] = b.builtinPWD
	
	// Job control builtins
	b.functions["jobs"] = b.builtinJobs
	b.functions["fg"] = b.builtinFG
	b.functions["bg"] = b.builtinBG
	b.functions["kill"] = b.builtinKill
	b.functions["wait"] = b.builtinWait
	
	// Utility builtins
	b.functions["date"] = b.builtinDate
	b.functions["sleep"] = b.builtinSleep
	b.functions["times"] = b.builtinTimes
	b.functions["umask"] = b.builtinUmask
	b.functions["ulimit"] = b.builtinUlimit
}

// GetBuiltin returns a builtin function if it exists
func (b *Builtins) GetBuiltin(name string) (BuiltinFunc, bool) {
	fn, exists := b.functions[name]
	return fn, exists
}

// Builtin implementations

func (b *Builtins) builtinEcho(ctx context.Context, env *Environment, args []string) (string, error) {
	// Parse echo options
	noNewline := false
	escape := false
	
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		switch args[i] {
		case "-n":
			noNewline = true
		case "-e":
			escape = true
		case "-E":
			escape = false
		default:
			break
		}
		i++
	}
	
	// Join remaining arguments
	output := strings.Join(args[i:], " ")
	
	if escape {
		output = b.escapeSequences(output)
	}
	
	if !noNewline {
		output += "\n"
	}
	
	return output, nil
}

func (b *Builtins) builtinPrintf(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	
	format := args[0]
	var formatArgs []interface{}
	
	// Convert remaining args to interface{}
	for _, arg := range args[1:] {
		formatArgs = append(formatArgs, arg)
	}
	
	result := fmt.Sprintf(format, formatArgs...)
	return result, nil
}

func (b *Builtins) builtinExport(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		// List all exported variables
		var output strings.Builder
		for name, value := range env.GetExportedVars() {
			output.WriteString(fmt.Sprintf("export %s=\"%s\"\n", name, value))
		}
		return output.String(), nil
	}
	
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Set and export variable
			parts := strings.SplitN(arg, "=", 2)
			if err := env.SetVar(parts[0], parts[1], true, false); err != nil {
				return "", err
			}
		} else {
			// Export existing variable
			if err := env.ExportVar(arg); err != nil {
				return "", err
			}
		}
	}
	
	return "", nil
}

func (b *Builtins) builtinUnset(ctx context.Context, env *Environment, args []string) (string, error) {
	for _, arg := range args {
		if err := env.UnsetVar(arg); err != nil {
			return "", err
		}
	}
	return "", nil
}

func (b *Builtins) builtinReadonly(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		// List all readonly variables
		var output strings.Builder
		env.mu.RLock()
		for name := range env.readonly {
			if value, exists := env.vars[name]; exists {
				output.WriteString(fmt.Sprintf("readonly %s=\"%s\"\n", name, value))
			}
		}
		env.mu.RUnlock()
		return output.String(), nil
	}
	
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if err := env.SetVar(parts[0], parts[1], false, true); err != nil {
				return "", err
			}
		} else {
			// Make existing variable readonly
			if err := env.SetVar(arg, "", false, true); err != nil {
				return "", err
			}
		}
	}
	
	return "", nil
}

func (b *Builtins) builtinSet(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		// List all variables
		var output strings.Builder
		env.mu.RLock()
		for name, value := range env.vars {
			output.WriteString(fmt.Sprintf("%s=\"%s\"\n", name, value))
		}
		env.mu.RUnlock()
		return output.String(), nil
	}
	
	// Parse options
	for _, arg := range args {
		if strings.HasPrefix(arg, "-o") {
			option := arg[2:]
			if option == "" {
				// List all options
				opts := env.GetOptions()
				output := fmt.Sprintf("allexport %t\n", opts.TrackAll)
				output += fmt.Sprintf("errexit %t\n", opts.ErrExit)
				output += fmt.Sprintf("ignoreeof %t\n", false) // TODO
				output += fmt.Sprintf("monitor %t\n", opts.Notify)
				output += fmt.Sprintf("noclobber %t\n", false) // TODO
				output += fmt.Sprintf("noglob %t\n", opts.NoGlob)
				output += fmt.Sprintf("noexec %t\n", opts.NoExec)
				output += fmt.Sprintf("noglob %t\n", opts.NoGlob)
				output += fmt.Sprintf("nounset %t\n", opts.NoUnset)
				output += fmt.Sprintf("verbose %t\n", opts.Verbose)
				output += fmt.Sprintf("xtrace %t\n", opts.XTrace)
				return output, nil
			} else {
				// Set option
				env.SetOption(option, true)
			}
		} else if strings.HasPrefix(arg, "+o") {
			option := arg[2:]
			env.SetOption(option, false)
		}
	}
	
	return "", nil
}

func (b *Builtins) builtinReturn(ctx context.Context, env *Environment, args []string) (string, error) {
	code := 0
	if len(args) > 0 {
		var err error
		code, err = strconv.Atoi(args[0])
		if err != nil {
			code = 1
		}
	}
	
	// TODO: Implement proper return handling
	return "", fmt.Errorf("return %d", code)
}

func (b *Builtins) builtinExit(ctx context.Context, env *Environment, args []string) (string, error) {
	code := 0
	if len(args) > 0 {
		var err error
		code, err = strconv.Atoi(args[0])
		if err != nil {
			code = 1
		}
	}
	
	// TODO: Implement proper exit handling
	return "", fmt.Errorf("exit %d", code)
}

func (b *Builtins) builtinBreak(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement break handling
	return "", fmt.Errorf("break")
}

func (b *Builtins) builtinContinue(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement continue handling
	return "", fmt.Errorf("continue")
}

func (b *Builtins) builtinColon(ctx context.Context, env *Environment, args []string) (string, error) {
	// : (colon) does nothing, always succeeds
	return "", nil
}

func (b *Builtins) builtinTrue(ctx context.Context, env *Environment, args []string) (string, error) {
	return "", nil
}

func (b *Builtins) builtinFalse(ctx context.Context, env *Environment, args []string) (string, error) {
	return "", fmt.Errorf("false")
}

func (b *Builtins) builtinAlias(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		// List all aliases
		var output strings.Builder
		for name, value := range env.aliases {
			output.WriteString(fmt.Sprintf("alias %s='%s'\n", name, value))
		}
		return output.String(), nil
	}
	
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			env.SetAlias(parts[0], parts[1])
		} else {
			// Show specific alias
			if value, exists := env.GetAlias(arg); exists {
				return fmt.Sprintf("alias %s='%s'\n", arg, value), nil
			}
		}
	}
	
	return "", nil
}

func (b *Builtins) builtinUnalias(ctx context.Context, env *Environment, args []string) (string, error) {
	for _, arg := range args {
		env.UnsetAlias(arg)
	}
	return "", nil
}

func (b *Builtins) builtinType(ctx context.Context, env *Environment, args []string) (string, error) {
	var output strings.Builder
	for _, arg := range args {
		if _, exists := env.GetFunction(arg); exists {
			output.WriteString(fmt.Sprintf("%s is a function\n", arg))
		} else if _, exists := env.GetAlias(arg); exists {
			output.WriteString(fmt.Sprintf("%s is an alias\n", arg))
		} else if _, exists := b.GetBuiltin(arg); exists {
			output.WriteString(fmt.Sprintf("%s is a shell builtin\n", arg))
		} else {
			output.WriteString(fmt.Sprintf("%s is not found\n", arg))
		}
	}
	return output.String(), nil
}

func (b *Builtins) builtinCommand(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement command lookup and execution
	return "", fmt.Errorf("command not implemented")
}

func (b *Builtins) builtinCD(ctx context.Context, env *Environment, args []string) (string, error) {
	dir := ""
	if len(args) > 0 {
		dir = args[0]
	} else {
		// Change to HOME directory
		if home, exists := env.GetVar("HOME"); exists {
			dir = home
		} else {
			return "", fmt.Errorf("cd: HOME not set")
		}
	}
	
	if dir == "-" {
		// Change to previous directory
		if oldpwd, exists := env.GetVar("OLDPWD"); exists {
			dir = oldpwd
		} else {
			return "", fmt.Errorf("cd: OLDPWD not set")
		}
	}
	
	// TODO: Implement actual directory change in K8s context
	env.SetVar("OLDPWD", env.GetWorkingDir(), false, false)
	env.SetWorkingDir(dir)
	env.SetVar("PWD", dir, true, false)
	
	return "", nil
}

func (b *Builtins) builtinPWD(ctx context.Context, env *Environment, args []string) (string, error) {
	return env.GetWorkingDir() + "\n", nil
}

func (b *Builtins) builtinJobs(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement job control
	return "", nil
}

func (b *Builtins) builtinFG(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement job control
	return "", fmt.Errorf("fg not implemented")
}

func (b *Builtins) builtinBG(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement job control
	return "", fmt.Errorf("bg not implemented")
}

func (b *Builtins) builtinKill(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement kill signal handling
	return "", fmt.Errorf("kill not implemented")
}

func (b *Builtins) builtinWait(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement job waiting
	return "", fmt.Errorf("wait not implemented")
}

func (b *Builtins) builtinDate(ctx context.Context, env *Environment, args []string) (string, error) {
	format := time.RFC3339
	if len(args) > 0 {
		format = args[0]
	}
	return time.Now().Format(format) + "\n", nil
}

func (b *Builtins) builtinSleep(ctx context.Context, env *Environment, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("sleep: missing operand")
	}
	
	duration, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", fmt.Errorf("sleep: invalid time interval")
	}
	
	select {
	case <-time.After(time.Duration(duration * float64(time.Second))):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	
	return "", nil
}

func (b *Builtins) builtinTimes(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement process timing
	return "0.00 0.00\n0.00 0.00\n", nil
}

func (b *Builtins) builtinUmask(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement umask handling
	return "", fmt.Errorf("umask not implemented")
}

func (b *Builtins) builtinUlimit(ctx context.Context, env *Environment, args []string) (string, error) {
	// TODO: Implement ulimit handling
	return "", fmt.Errorf("ulimit not implemented")
}

// Helper functions

func (b *Builtins) escapeSequences(input string) string {
	// Handle common escape sequences
	replacements := map[string]string{
		"\\a": "\a",
		"\\b": "\b",
		"\\f": "\f",
		"\\n": "\n",
		"\\r": "\r",
		"\\t": "\t",
		"\\v": "\v",
		"\\\\": "\\",
		"\\'": "'",
		"\\\"": "\"",
	}
	
	result := input
	for seq, replacement := range replacements {
		result = strings.ReplaceAll(result, seq, replacement)
	}
	
	return result
}
