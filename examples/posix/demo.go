package main

import (
	"context"
	"fmt"

	"github.com/rmasci/k8sh/pkg/posix"
)

func main() {
	fmt.Println("🐚 k8sh POSIX Shell Demo")
	fmt.Println("========================")
	
	// Create a standalone POSIX environment (no Kubernetes needed)
	env := posix.NewEnvironment()
	executor := posix.NewExecutor(env)
	ctx := context.Background()
	
	// Demo commands to showcase POSIX features
	demoCommands := []string{
		"echo 'Hello from POSIX shell!'",
		"printf 'Simple output: %s\n' 'test'",
		"export DEMO_VAR=posix_demo",
		"echo $DEMO_VAR",
		"pwd",
		"cd /tmp",
		"pwd",
		"help",
	}
	
	fmt.Println("Running POSIX command demonstrations...")
	fmt.Println()
	
	for i, cmd := range demoCommands {
		fmt.Printf("=== Command %d: %s ===\n", i+1, cmd)
		
		// Parse the command
		parser := posix.NewParser(cmd)
		ast, err := parser.Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}
		
		fmt.Printf("AST: %s\n", ast.String())
		
		// Execute the command
		result, err := executor.Execute(ctx, ast)
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
		} else if result != "" {
			fmt.Printf("Output: %s", result)
		}
		
		fmt.Println()
	}
	
	// Show environment state
	fmt.Println("=== Environment State ===")
	exportedVars := env.GetExportedVars()
	fmt.Println("Exported variables:")
	for name, value := range exportedVars {
		fmt.Printf("  %s=%s\n", name, value)
	}
	
	fmt.Printf("Current working directory: %s\n", env.GetWorkingDir())
	
	fmt.Println()
	fmt.Println("🎉 POSIX Shell Demo Complete!")
	fmt.Println("To use with Kubernetes, run: ./bin/k8sh posix")
}
