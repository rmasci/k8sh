package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/rest"
)

type FileOperations struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewFileOperations(clientset *kubernetes.Clientset, config *rest.Config) *FileOperations {
	return &FileOperations{
		clientset: clientset,
		config:    config,
	}
}

// ListDirectory lists contents of a directory in a container
func (fo *FileOperations) ListDirectory(ctx context.Context, namespace, pod, container, path string) ([]string, error) {
	// Use ls command to list directory
	cmd := []string{"ls", "-la", path}
	
	req := fo.clientset.CoreV1().RESTClient().Post().
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

	executor, err := remotecommand.NewSPDYExecutor(fo.config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	var output strings.Builder
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &output,
		Stderr: os.Stderr,
	})

	if err != nil {
		// Try with ephemeral container if direct exec fails
		return fo.listDirectoryViaEphemeral(ctx, namespace, pod, container, path)
	}

	lines := strings.Split(output.String(), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// ReadFile reads a file from a container
func (fo *FileOperations) ReadFile(ctx context.Context, namespace, pod, container, path string) (string, error) {
	cmd := []string{"cat", path}
	
	req := fo.clientset.CoreV1().RESTClient().Post().
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

	executor, err := remotecommand.NewSPDYExecutor(fo.config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var output strings.Builder
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &output,
		Stderr: os.Stderr,
	})

	if err != nil {
		// Try with ephemeral container if direct exec fails
		return fo.readFileViaEphemeral(ctx, namespace, pod, container, path)
	}

	return output.String(), nil
}

// WriteFile writes content to a file in a container
func (fo *FileOperations) WriteFile(ctx context.Context, namespace, pod, container, path, content string) error {
	cmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' > %s", content, path)}
	
	req := fo.clientset.CoreV1().RESTClient().Post().
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

	executor, err := remotecommand.NewSPDYExecutor(fo.config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})

	if err != nil {
		// Try with ephemeral container if direct exec fails
		return fo.writeFileViaEphemeral(ctx, namespace, pod, container, path, content)
	}

	return nil
}

// Fallback methods for distroless containers
func (fo *FileOperations) listDirectoryViaEphemeral(ctx context.Context, namespace, pod, container, path string) ([]string, error) {
	// TODO: Implement ephemeral container approach
	return []string{"TODO: Implement ephemeral container directory listing"}, nil
}

func (fo *FileOperations) readFileViaEphemeral(ctx context.Context, namespace, pod, container, path string) (string, error) {
	// TODO: Implement ephemeral container file reading
	return "TODO: Implement ephemeral container file reading", nil
}

func (fo *FileOperations) writeFileViaEphemeral(ctx context.Context, namespace, pod, container, path, content string) error {
	// TODO: Implement ephemeral container file writing
	return fmt.Errorf("TODO: Implement ephemeral container file writing")
}
