package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/posix"
	"github.com/rmasci/k8sh/pkg/shell"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "k8sh",
		Short: "A pseudo-shell for Kubernetes pods",
		Long: `k8sh is an OS-independent pseudo-shell for Kubernetes pods that works
without requiring any tools in target containers. Supports distroless,
scratch, alpine, debian, and ubuntu-based images.`,
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

			// Start interactive shell
			if err := shell.StartShell(client); err != nil {
				fmt.Printf("Error starting shell: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add POSIX subcommand
	var posixCmd = &cobra.Command{
		Use:   "posix",
		Short: "Start POSIX-compliant shell",
		Long: `Start a POSIX-compliant shell with full command parsing,
pipelines, redirection, and built-in commands.`,
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

			// Start POSIX shell
			if err := posix.StartPOSIXShell(client); err != nil {
				fmt.Printf("Error starting POSIX shell: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add subcommand to root
	rootCmd.AddCommand(posixCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
