package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/posix"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "k8sh",
		Short: "A POSIX-compliant pseudo-shell for Kubernetes pods",
		Long: `k8sh is a POSIX-compliant pseudo-shell for Kubernetes pods that works
without requiring any tools in target containers. Supports distroless,
scratch, alpine, debian, and ubuntu-based images.

🐚 POSIX SHELL FEATURES:
  • Command pipelines: cmd1 | cmd2 | cmd3
  • I/O redirection: >, >>, <, 2>, &>
  • Variable expansion: $VAR, ${VAR}
  • Command substitution: $(cmd)
  • Built-in commands: echo, printf, export, cd, pwd, help
  • Environment management: export, unset, readonly
  • Shell options and error handling
  • Tab completion and history navigation

🚀 QUICK START:
  k8sh                    # Start POSIX shell
  k8sh --help             # Show this help
  k8sh version            # Show version info

📁 FEATURES:
  • File operations: ls, cd, cat, vi, mkdir, rm, cp, mv, touch
  • Text processing: head, tail, grep, wc, sort
  • System information: ps, env, df, du, ip
  • Kubernetes: pods, use, namespace
  • Download files: download <src> <dst>
  • POSIX compliance: Full shell with pipelines and redirection

🎯 EXAMPLES:
  # Inside k8sh shell:
  echo "Hello POSIX!"           # Basic command
  export MY_VAR=value           # Set variable
  echo $MY_VAR                  # Variable expansion
  cat file.txt | grep "error"   # Pipeline
  ls -la > files.txt            # Redirection
  use my-pod                    # Select Kubernetes pod
  download /app/logs/app.log ./backup.log  # Download file

⚙️  CONFIGURATION:
  Uses standard Kubernetes config at ~/.kube/config
  First-time setup: minikube start, gcloud init, or manual config

📚 POSIX COMPLIANCE:
  Full POSIX shell compliance with:
  • AST-based parsing and execution
  • Environment and variable management
  • Built-in command implementation
  • Pipeline and redirection support
  • Job control and process management

The k8sh shell provides a complete Unix-like shell experience
in Kubernetes containers without requiring any tools in the target
container. Perfect for distroless and minimal environments!`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := k8s.NewClient()
			if err != nil {
				// Check if this is a config error
				if strings.Contains(err.Error(), "kube/config") || strings.Contains(err.Error(), "no such file") {
					fmt.Printf(`🔍 KUBERNETES CONFIG NOT FOUND

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

After setup, run k8sh again! 🎉

`, clientcmd.RecommendedHomeFile, filepath.Dir(clientcmd.RecommendedHomeFile))
				} else {
					fmt.Printf("Error creating Kubernetes client: %v\n", err)
				}
				os.Exit(1)
			}

			// Start POSIX shell by default
			if err := posix.StartPOSIXShell(client); err != nil {
				fmt.Printf("Error starting POSIX shell: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Display version information for k8sh including build details,
supported platforms, and feature status.

📊 VERSION INFORMATION:
  • Current version and build details
  • Supported platforms and architectures
  • Feature availability and status
  • Kubernetes compatibility information

🎯 USAGE:
  k8sh version                 # Show version info
  k8sh version --help          # Show this help

This command helps you verify your k8sh installation and
check which features are available in your build.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`🐚 k8sh - Kubernetes Pseudo-Shell
=====================================
Version: 1.0.0
Build:   Cross-platform distribution
Platform: %s/%s
Go:      %s

🎯 SUPPORTED PLATFORMS:
  • macOS (Intel, Apple Silicon)
  • Linux (Intel, ARM)
  • Windows (Intel)

📁 FEATURES:
  ✅ POSIX-compliant shell (pipelines, redirection, variables)
  ✅ File operations (ls, cd, cat, vi, mkdir, rm, cp, mv, touch)
  ✅ Text processing (grep, wc, sort, head, tail)
  ✅ System information (ps, env, df, du, ip)
  ✅ Kubernetes integration (pods, use, namespace)
  ✅ File download (download command)
  ✅ Tab completion and history navigation
  ✅ Cross-platform builds

⚙️  KUBERNETES:
  • Standard ~/.kube/config support
  • Multiple namespace support
  • Container selection
  • Ephemeral container support

📦 BUILT FOR DISTRIBUTION:
  • No container dependencies
  • Works in distroless containers
  • Single binary deployment
  • Professional POSIX shell experience

🐚 POSIX COMPLIANCE:
  Full POSIX shell compliance with:
  • AST-based parsing and execution
  • Environment and variable management
  • Built-in command implementation
  • Pipeline and redirection support
  • Job control and process management

For more information: https://github.com/rmasci/k8sh
`, runtime.GOOS, runtime.GOARCH, runtime.Version())
		},
	}

	// Add completion command
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `Generate shell completion scripts for k8sh to enable
tab completion in your preferred shell.

🎯 USAGE:
  k8sh completion bash        # Generate bash completion
  k8sh completion zsh         # Generate zsh completion
  k8sh completion fish        # Generate fish completion
  k8sh completion powershell # Generate PowerShell completion

📦 INSTALLATION:
  # Bash
  k8sh completion bash > ~/.k8sh-completion.bash
  echo 'source ~/.k8sh-completion.bash' >> ~/.bashrc

  # Zsh
  k8sh completion zsh > ~/.k8sh-completion.zsh
  echo 'source ~/.k8sh-completion.zsh' >> ~/.zshrc

  # Fish
  k8sh completion fish > ~/.config/fish/completions/k8sh.fish

This enables tab completion for k8sh commands and options
in your terminal!`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Please specify a shell: bash, zsh, fish, or powershell")
				fmt.Println("Example: k8sh completion bash")
				return
			}
			
			shell := args[0]
			switch shell {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				fmt.Printf("Unsupported shell: %s\n", shell)
				fmt.Println("Supported shells: bash, zsh, fish, powershell")
			}
		},
	}

	// Add subcommands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
