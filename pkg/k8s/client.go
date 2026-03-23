package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewClient() (*Client, error) {
	// Use kubeconfig from default location
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// DownloadFile downloads a file or directory from a pod to local filesystem
// This implementation doesn't depend on tar being available in the container
func (c *Client) DownloadFile(namespace, pod, container, srcPath, dstPath string, recursive bool) error {
	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Check if source is a file or directory
	isDir, err := c.isRemoteDirectory(namespace, pod, container, srcPath)
	if err != nil {
		return fmt.Errorf("failed to check source type: %w", err)
	}

	if isDir && recursive {
		// Download directory recursively
		return c.downloadDirectory(namespace, pod, container, srcPath, dstPath)
	} else if isDir {
		// Download single directory contents
		return c.downloadDirectoryContents(namespace, pod, container, srcPath, dstPath)
	} else {
		// Download single file
		return c.downloadSingleFile(namespace, pod, container, srcPath, dstPath)
	}
}

// isRemoteDirectory checks if a path is a directory in the remote pod
func (c *Client) isRemoteDirectory(namespace, pod, container, path string) (bool, error) {
	// Use 'test -d' command to check if path is a directory
	cmd := []string{"test", "-d", path}
	
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, nil)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return false, err
	}

	// Capture output to check if it's a directory
	var output strings.Builder
	err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &output,
		Stderr: os.Stderr,
	})
	
	if err != nil {
		return false, err
	}

	// test -d returns 0 if it's a directory, 1 if it's a file
	return strings.TrimSpace(output.String()) == "0", nil
}

// downloadSingleFile downloads a single file from pod to local
func (c *Client) downloadSingleFile(namespace, pod, container, srcPath, dstPath string) error {
	// Use 'cat' command to read file content
	cmd := []string{"cat", srcPath}
	
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, nil)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	// Create destination file
	file, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close()

	// Stream content directly to file
	err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: file,
		Stderr: os.Stderr,
	})
	
	if err != nil {
		os.Remove(dstPath) // Clean up on error
		return fmt.Errorf("failed to download file content: %w", err)
	}

	return nil
}

// downloadDirectoryContents downloads all files in a directory (non-recursive)
func (c *Client) downloadDirectoryContents(namespace, pod, container, srcPath, dstPath string) error {
	// Use 'find' command to list files (basic implementation)
	cmd := []string{"sh", "-c", fmt.Sprintf("find %s -maxdepth 1 -type f", srcPath)}
	
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, nil)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	// Get file list
	var output strings.Builder
	err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &output,
		Stderr: os.Stderr,
	})
	
	if err != nil {
		return fmt.Errorf("failed to list directory contents: %w", err)
	}

	// Download each file
	files := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, file := range files {
		if file == "" {
			continue
		}
		
		// Extract filename from full path
		filename := filepath.Base(file)
		localPath := filepath.Join(dstPath, filename)
		
		// Download the file
		err := c.downloadSingleFile(namespace, pod, container, file, localPath)
		if err != nil {
			fmt.Printf("Warning: failed to download %s: %v\n", filename, err)
		}
	}

	return nil
}

// downloadDirectory downloads a directory recursively
func (c *Client) downloadDirectory(namespace, pod, container, srcPath, dstPath string) error {
	// Use 'find' command to list all files recursively
	cmd := []string{"sh", "-c", fmt.Sprintf("find %s -type f", srcPath)}
	
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, nil)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	// Get file list
	var output strings.Builder
	err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &output,
		Stderr: os.Stderr,
	})
	
	if err != nil {
		return fmt.Errorf("failed to list directory contents: %w", err)
	}

	// Download each file, preserving directory structure
	files := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, file := range files {
		if file == "" {
			continue
		}
		
		// Calculate relative path and local path
		relPath, err := filepath.Rel(srcPath, file)
		if err != nil {
			relPath = filepath.Base(file)
		}
		
		localPath := filepath.Join(dstPath, relPath)
		
		// Create local directory if needed
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			fmt.Printf("Warning: failed to create directory %s: %v\n", localDir, err)
		}
		
		// Download the file
		err = c.downloadSingleFile(namespace, pod, container, file, localPath)
		if err != nil {
			fmt.Printf("Warning: failed to download %s: %v\n", file, err)
		}
	}

	return nil
}
