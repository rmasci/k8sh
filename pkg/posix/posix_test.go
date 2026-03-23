package posix

import (
	"context"
	"testing"
)

func TestParser(t *testing.T) {
	t.Run("SimpleCommand", func(t *testing.T) {
		parser := NewParser("echo hello world")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		cmd, ok := ast.(*CommandNode)
		if !ok {
			t.Fatalf("Expected CommandNode, got %T", ast)
		}
		
		if cmd.Command != "echo" {
			t.Errorf("Expected command 'echo', got '%s'", cmd.Command)
		}
		
		if len(cmd.Args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(cmd.Args))
		}
	})
	
	t.Run("QuotedString", func(t *testing.T) {
		parser := NewParser("echo \"hello world\"")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		cmd, ok := ast.(*CommandNode)
		if !ok {
			t.Fatalf("Expected CommandNode, got %T", ast)
		}
		
		if len(cmd.Args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(cmd.Args))
		}
		
		arg, ok := cmd.Args[0].(*StringNode)
		if !ok {
			t.Fatalf("Expected StringNode, got %T", cmd.Args[0])
		}
		
		if arg.Value != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", arg.Value)
		}
	})
	
	t.Run("VariableExpansion", func(t *testing.T) {
		parser := NewParser("echo $PATH")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		cmd, ok := ast.(*CommandNode)
		if !ok {
			t.Fatalf("Expected CommandNode, got %T", ast)
		}
		
		if len(cmd.Args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(cmd.Args))
		}
		
		arg, ok := cmd.Args[0].(*VariableNode)
		if !ok {
			t.Fatalf("Expected VariableNode, got %T", cmd.Args[0])
		}
		
		if arg.Name != "PATH" {
			t.Errorf("Expected 'PATH', got '%s'", arg.Name)
		}
	})
	
	t.Run("CommandSubstitution", func(t *testing.T) {
		parser := NewParser("echo $(date)")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		cmd, ok := ast.(*CommandNode)
		if !ok {
			t.Fatalf("Expected CommandNode, got %T", ast)
		}
		
		if len(cmd.Args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(cmd.Args))
		}
		
		_, ok = cmd.Args[0].(*SubstitutionNode)
		if !ok {
			t.Fatalf("Expected SubstitutionNode, got %T", cmd.Args[0])
		}
	})
	
	t.Run("Redirection", func(t *testing.T) {
		parser := NewParser("echo hello > file.txt")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		cmd, ok := ast.(*CommandNode)
		if !ok {
			t.Fatalf("Expected CommandNode, got %T", ast)
		}
		
		if len(cmd.Redirs) != 1 {
			t.Fatalf("Expected 1 redirection, got %d", len(cmd.Redirs))
		}
		
		redir := cmd.Redirs[0]
		if redir.Type != RedirOutput {
			t.Errorf("Expected RedirOutput, got %v", redir.Type)
		}
		
		if redir.Target != "file.txt" {
			t.Errorf("Expected 'file.txt', got '%s'", redir.Target)
		}
	})
	
	t.Run("Pipeline", func(t *testing.T) {
		parser := NewParser("cat file.txt | grep hello")
		ast, err := parser.Parse()
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		pipeline, ok := ast.(*PipelineNode)
		if !ok {
			t.Fatalf("Expected PipelineNode, got %T", ast)
		}
		
		if len(pipeline.Commands) != 2 {
			t.Fatalf("Expected 2 commands, got %d", len(pipeline.Commands))
		}
	})
}

func TestEnvironment(t *testing.T) {
	env := NewEnvironment()
	
	t.Run("SetAndGetVar", func(t *testing.T) {
		err := env.SetVar("TEST_VAR", "test_value", false, false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		value, exists := env.GetVar("TEST_VAR")
		if !exists {
			t.Error("Expected variable to exist")
		}
		
		if value != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", value)
		}
	})
	
	t.Run("ExportVar", func(t *testing.T) {
		err := env.SetVar("EXPORT_VAR", "export_value", true, false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		exported := env.GetExportedVars()
		if _, exists := exported["EXPORT_VAR"]; !exists {
			t.Error("Expected variable to be exported")
		}
	})
	
	t.Run("ReadonlyVar", func(t *testing.T) {
		err := env.SetVar("READONLY_VAR", "readonly_value", false, true)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Try to modify readonly variable
		err = env.SetVar("READONLY_VAR", "new_value", false, false)
		if err == nil {
			t.Error("Expected error when modifying readonly variable")
		}
	})
	
	t.Run("UnsetVar", func(t *testing.T) {
		err := env.SetVar("UNSET_VAR", "unset_value", false, false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		err = env.UnsetVar("UNSET_VAR")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		_, exists := env.GetVar("UNSET_VAR")
		if exists {
			t.Error("Expected variable to be unset")
		}
	})
	
	t.Run("ExpandVariables", func(t *testing.T) {
		env.SetVar("NAME", "world", false, false)
		
		result := env.ExpandVariables("Hello $NAME!")
		if result != "Hello world!" {
			t.Errorf("Expected 'Hello world!', got '%s'", result)
		}
	})
	
	t.Run("WorkingDir", func(t *testing.T) {
		env.SetWorkingDir("/tmp")
		
		if env.GetWorkingDir() != "/tmp" {
			t.Errorf("Expected '/tmp', got '%s'", env.GetWorkingDir())
		}
		
		pwd, exists := env.GetVar("PWD")
		if !exists {
			t.Error("Expected PWD to be set")
		}
		
		if pwd != "/tmp" {
			t.Errorf("Expected PWD to be '/tmp', got '%s'", pwd)
		}
	})
}

func TestBuiltins(t *testing.T) {
	ctx := context.Background()
	env := NewEnvironment()
	builtins := NewBuiltins()
	
	t.Run("Builtins", func(t *testing.T) {
		echoFn, exists := builtins.GetBuiltin("echo")
		if !exists {
			t.Fatal("Expected echo builtin to exist")
		}
		
		result, err := echoFn(ctx, env, []string{"hello", "world"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		expected := "hello world\n"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("Printf", func(t *testing.T) {
		printfFn, exists := builtins.GetBuiltin("printf")
		if !exists {
			t.Fatal("Expected printf builtin to exist")
		}
		
		result, err := printfFn(ctx, env, []string{"Hello %s", "world"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		expected := "Hello world"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("Export", func(t *testing.T) {
		exportFn, exists := builtins.GetBuiltin("export")
		if !exists {
			t.Fatal("Expected export builtin to exist")
		}
		
		result, err := exportFn(ctx, env, []string{"TEST_VAR=test_value"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
		
		value, exists := env.GetVar("TEST_VAR")
		if !exists {
			t.Error("Expected TEST_VAR to be set")
		}
		
		if value != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", value)
		}
		
		exported := env.GetExportedVars()
		if _, exists := exported["TEST_VAR"]; !exists {
			t.Error("Expected TEST_VAR to be exported")
		}
	})
	
	t.Run("True", func(t *testing.T) {
		trueFn, exists := builtins.GetBuiltin("true")
		if !exists {
			t.Fatal("Expected true builtin to exist")
		}
		
		result, err := trueFn(ctx, env, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
	
	t.Run("False", func(t *testing.T) {
		falseFn, exists := builtins.GetBuiltin("false")
		if !exists {
			t.Fatal("Expected false builtin to exist")
		}
		
		result, err := falseFn(ctx, env, []string{})
		if err == nil {
			t.Error("Expected error from false builtin")
		}
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
}

func TestExecutor(t *testing.T) {
	ctx := context.Background()
	env := NewEnvironment()
	executor := NewExecutor(env)
	
	t.Run("ExecuteStringNode", func(t *testing.T) {
		node := &StringNode{Value: "test"}
		result, err := executor.Execute(ctx, node)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "test" {
			t.Errorf("Expected 'test', got '%s'", result)
		}
	})
	
	t.Run("ExecuteVariableNode", func(t *testing.T) {
		env.SetVar("TEST_VAR", "test_value", false, false)
		node := &VariableNode{Name: "TEST_VAR"}
		result, err := executor.Execute(ctx, node)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}
	})
	
	t.Run("ExecuteSimpleCommand", func(t *testing.T) {
		node := &CommandNode{
			Command: "echo",
			Args: []ASTNode{
				&StringNode{Value: "hello"},
				&StringNode{Value: "world"},
			},
		}
		
		result, err := executor.Execute(ctx, node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		expected := "hello world\n"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("ExecuteBuiltinCommand", func(t *testing.T) {
		node := &CommandNode{
			Command: "pwd",
			Args:    []ASTNode{},
		}
		
		result, err := executor.Execute(ctx, node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "/\n" {
			t.Errorf("Expected '/\\n', got '%s'", result)
		}
	})
}

func TestIntegration(t *testing.T) {
	// Note: These tests would require mocking the k8s shell
	// For now, we'll test the POSIX components in isolation
	
	t.Run("ParserIntegration", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"echo hello", "hello\n"},
			{"echo -n hello", "hello"},
			{"printf 'Hello %s' world", "Hello world"},
		}
		
		ctx := context.Background()
		env := NewEnvironment()
		executor := NewExecutor(env)
		
		for _, tc := range testCases {
			parser := NewParser(tc.input)
			ast, err := parser.Parse()
			if err != nil {
				t.Errorf("Parse error for '%s': %v", tc.input, err)
				continue
			}
			
			result, err := executor.Execute(ctx, ast)
			if err != nil {
				t.Errorf("Execute error for '%s': %v", tc.input, err)
				continue
			}
			
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s' for '%s'", tc.expected, result, tc.input)
			}
		}
	})
}

func TestASTNodes(t *testing.T) {
	t.Run("StringNode", func(t *testing.T) {
		node := &StringNode{Value: "test"}
		if node.String() != `"test"` {
			t.Errorf("Expected '\"test\"', got '%s'", node.String())
		}
	})
	
	t.Run("VariableNode", func(t *testing.T) {
		node := &VariableNode{Name: "PATH"}
		if node.String() != "$PATH" {
			t.Errorf("Expected '$PATH', got '%s'", node.String())
		}
	})
	
	t.Run("CommandNode", func(t *testing.T) {
		node := &CommandNode{
			Command: "echo",
			Args: []ASTNode{
				&StringNode{Value: "hello"},
			},
		}
		expected := `echo "hello"`
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})
	
	t.Run("PipelineNode", func(t *testing.T) {
		node := &PipelineNode{
			Commands: []ASTNode{
				&CommandNode{Command: "cat", Args: []ASTNode{&StringNode{Value: "file.txt"}}},
				&CommandNode{Command: "grep", Args: []ASTNode{&StringNode{Value: "pattern"}}},
			},
		}
		expected := `cat "file.txt" | grep "pattern"`
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})
}
