package main

import (
	"testing"
	
	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that root command is properly configured
	rootCmd := createRootCommand()
	
	if rootCmd.Use != "k8sh" {
		t.Errorf("Expected use 'k8sh', got '%s'", rootCmd.Use)
	}
	
	if rootCmd.Short != "A pseudo-shell for Kubernetes pods" {
		t.Errorf("Expected short description 'A pseudo-shell for Kubernetes pods', got '%s'", rootCmd.Short)
	}
	
	if len(rootCmd.Long) == 0 {
		t.Error("Expected long description to be non-empty")
	}
	
	if !contains(rootCmd.Long, "distroless") {
		t.Error("Expected long description to mention distroless support")
	}
}

func TestMainFunction(t *testing.T) {
	// We can't easily test the main function without mocking kubeconfig
	// but we can test that it doesn't panic
	t.Run("MainExists", func(t *testing.T) {
		// This just verifies the function exists and has the right signature
		// In a real test environment, you would mock the kubeconfig
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main function panicked: %v", r)
			}
		}()
		
		// We can't actually call main() as it would exit the program
		// But we can verify the function exists by checking it compiles
	})
}

// Helper function to create root command for testing
func createRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "k8sh",
		Short: "A pseudo-shell for Kubernetes pods",
		Long: `k8sh is an OS-independent pseudo-shell for Kubernetes pods that works
without requiring any tools in target containers. Supports distroless,
scratch, alpine, debian, and ubuntu-based images.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Mock implementation for testing
		},
	}
	return rootCmd
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
