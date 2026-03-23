package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
)

// Essential file tools
func (fo *FileOperations) MakeDirectory(ctx context.Context, namespace, pod, container, path string) error {
	cmd := []string{"mkdir", "-p", path}
	
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

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})
}

func (fo *FileOperations) RemoveFile(ctx context.Context, namespace, pod, container, path string, recursive bool) error {
	cmd := []string{"rm"}
	if recursive {
		cmd = append(cmd, "-r")
	}
	cmd = append(cmd, path)
	
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

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})
}

func (fo *FileOperations) CopyFile(ctx context.Context, namespace, pod, container, src, dst string) error {
	cmd := []string{"cp", src, dst}
	
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

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})
}

func (fo *FileOperations) MoveFile(ctx context.Context, namespace, pod, container, src, dst string) error {
	cmd := []string{"mv", src, dst}
	
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

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})
}

func (fo *FileOperations) TouchFile(ctx context.Context, namespace, pod, container, path string) error {
	cmd := []string{"touch", path}
	
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

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: os.Stderr,
	})
}

// Text processing tools
func (fo *FileOperations) HeadFile(ctx context.Context, namespace, pod, container, path string, lines int) (string, error) {
	cmd := []string{"head"}
	if lines > 0 {
		cmd = append(cmd, "-n", strconv.Itoa(lines))
	}
	cmd = append(cmd, path)
	
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

	return output.String(), err
}

func (fo *FileOperations) TailFile(ctx context.Context, namespace, pod, container, path string, lines int) (string, error) {
	cmd := []string{"tail"}
	if lines > 0 {
		cmd = append(cmd, "-n", strconv.Itoa(lines))
	}
	cmd = append(cmd, path)
	
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

	return output.String(), err
}

func (fo *FileOperations) GrepFile(ctx context.Context, namespace, pod, container, pattern, path string, ignoreCase bool) (string, error) {
	cmd := []string{"grep"}
	if ignoreCase {
		cmd = append(cmd, "-i")
	}
	cmd = append(cmd, pattern, path)
	
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

	return output.String(), err
}

func (fo *FileOperations) WordCount(ctx context.Context, namespace, pod, container, path string) (string, error) {
	cmd := []string{"wc", path}
	
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

	return output.String(), err
}

func (fo *FileOperations) SortFile(ctx context.Context, namespace, pod, container, path string, unique bool) (string, error) {
	cmd := []string{"sort"}
	if unique {
		cmd = append(cmd, "-u")
	}
	cmd = append(cmd, path)
	
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

	return output.String(), err
}

// System information tools
func (fo *FileOperations) ListProcesses(ctx context.Context, namespace, pod, container string) (string, error) {
	cmd := []string{"ps", "aux"}
	
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

	return output.String(), err
}

func (fo *FileOperations) GetEnvironment(ctx context.Context, namespace, pod, container string) (string, error) {
	cmd := []string{"env"}
	
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

	return output.String(), err
}

func (fo *FileOperations) GetDiskUsage(ctx context.Context, namespace, pod, container, path string) (string, error) {
	cmd := []string{"df", "-h", path}
	
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

	return output.String(), err
}

func (fo *FileOperations) GetDirectoryUsage(ctx context.Context, namespace, pod, container, path string) (string, error) {
	cmd := []string{"du", "-sh", path}
	
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

	return output.String(), err
}

// Local text processing for when container tools aren't available
func GrepLocal(content, pattern string, ignoreCase bool) string {
	var flags string
	if ignoreCase {
		flags = "(?i)"
	}
	
	regex, err := regexp.Compile(flags + pattern)
	if err != nil {
		return fmt.Sprintf("Invalid regex: %v", err)
	}

	lines := strings.Split(content, "\n")
	var matches []string
	
	for _, line := range lines {
		if regex.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return strings.Join(matches, "\n")
}

func WordCountLocal(content string) string {
	lines := strings.Split(content, "\n")
	words := strings.Fields(content)
	chars := len(content)
	
	return fmt.Sprintf("%d %d %d", len(lines), len(words), chars)
}

func SortLocal(content string, unique bool) string {
	lines := strings.Split(content, "\n")
	
	// Remove empty lines
	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	
	// Sort
	sort.Strings(nonEmpty)
	
	// Remove duplicates if unique flag is set
	if unique {
		var uniqueLines []string
		seen := make(map[string]bool)
		for _, line := range nonEmpty {
			if !seen[line] {
				seen[line] = true
				uniqueLines = append(uniqueLines, line)
			}
		}
		nonEmpty = uniqueLines
	}
	
	return strings.Join(nonEmpty, "\n")
}

func HeadLocal(content string, lines int) string {
	contentLines := strings.Split(content, "\n")
	if lines <= 0 || lines >= len(contentLines) {
		return content
	}
	
	return strings.Join(contentLines[:lines], "\n")
}

func TailLocal(content string, lines int) string {
	contentLines := strings.Split(content, "\n")
	if lines <= 0 || lines >= len(contentLines) {
		return content
	}
	
	start := len(contentLines) - lines
	return strings.Join(contentLines[start:], "\n")
}
