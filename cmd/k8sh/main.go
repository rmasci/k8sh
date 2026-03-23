package main

import (
	"fmt"
	"os"

	"github.com/rmasci/k8sh/pkg/k8s"
	"github.com/rmasci/k8sh/pkg/shell"
	"github.com/spf13/cobra"
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
				fmt.Printf("Error creating Kubernetes client: %v\n", err)
				os.Exit(1)
			}

			// Start interactive shell
			if err := shell.StartShell(client); err != nil {
				fmt.Printf("Error starting shell: %v\n", err)
				os.Exit(1)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
