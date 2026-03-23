package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodInfo struct {
	Name       string
	Namespace  string
	Containers []ContainerInfo
	Status     string
	Node       string
}

type ContainerInfo struct {
	Name  string
	Image string
	Ready bool
}

func (c *Client) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		var containers []ContainerInfo
		for _, container := range pod.Spec.Containers {
			ready := false
			for _, status := range pod.Status.ContainerStatuses {
				if status.Name == container.Name {
					ready = status.Ready
					break
				}
			}
			containers = append(containers, ContainerInfo{
				Name:  container.Name,
				Image: container.Image,
				Ready: ready,
			})
		}

		podInfos = append(podInfos, PodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Containers: containers,
			Status:     string(pod.Status.Phase),
			Node:       pod.Spec.NodeName,
		})
	}

	return podInfos, nil
}

func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

func (c *Client) CreateEphemeralContainer(ctx context.Context, namespace, podName, targetContainer string) error {
	// Create ephemeral container spec
	ephemeral := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  "k8sh-debug",
			Image: "alpine:latest",
			Command: []string{"/bin/sh"},
			Stdin: true,
			TTY:   true,
			SecurityContext: &corev1.SecurityContext{
				Privileged: &[]bool{true}[0],
			},
		},
		TargetContainerName: targetContainer,
	}

	// Patch the pod to add ephemeral container
	patch := fmt.Sprintf(`{
		"spec": {
			"ephemeralContainers": [%s]
		}
	}`, ephemeralToString(ephemeral))

	_, err := c.clientset.CoreV1().Pods(namespace).Patch(
		ctx,
		podName,
		"application/json-patch+json",
		[]byte(patch),
		metav1.PatchOptions{},
	)

	return err
}

func ephemeralToString(ec corev1.EphemeralContainer) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(`"name": "%s"`, ec.Name))
	parts = append(parts, fmt.Sprintf(`"image": "%s"`, ec.Image))
	parts = append(parts, fmt.Sprintf(`"command": [%s]`, arrayToString(ec.Command)))
	parts = append(parts, fmt.Sprintf(`"stdin": %t`, ec.Stdin))
	parts = append(parts, fmt.Sprintf(`"tty": %t`, ec.TTY))
	
	if ec.SecurityContext != nil && ec.SecurityContext.Privileged != nil {
		parts = append(parts, fmt.Sprintf(`"securityContext": {"privileged": %t}`, *ec.SecurityContext.Privileged))
	}
	
	if ec.TargetContainerName != "" {
		parts = append(parts, fmt.Sprintf(`"targetContainerName": "%s"`, ec.TargetContainerName))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

func arrayToString(arr []string) string {
	var quoted []string
	for _, s := range arr {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, s))
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
