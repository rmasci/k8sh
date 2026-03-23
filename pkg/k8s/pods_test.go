package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"

	k8stesting "github.com/rmasci/k8sh/pkg/testing"
)

func TestListPods(t *testing.T) {
	ctx := context.Background()
	
	t.Run("EmptyNamespace", func(t *testing.T) {
		clientset := k8stesting.NewFakeKubernetesClient()
		client := &Client{clientset: clientset, config: k8stesting.GetTestConfig()}
		
		pods, err := client.ListPods(ctx, k8stesting.TestNamespace)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if len(pods) != 0 {
			t.Errorf("Expected 0 pods, got %d", len(pods))
		}
	})
	
	t.Run("WithPods", func(t *testing.T) {
		// Create test pods
		pod1 := testing.CreateTestPod("pod1", testing.TestNamespace)
		pod2 := testing.CreateTestPod("pod2", testing.TestNamespace)
		
		clientset := testing.NewFakeKubernetesClient(pod1, pod2)
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		pods, err := client.ListPods(ctx, testing.TestNamespace)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if len(pods) != 2 {
			t.Errorf("Expected 2 pods, got %d", len(pods))
		}
		
		// Check pod details
		foundPod1 := false
		foundPod2 := false
		for _, pod := range pods {
			if pod.Name == "pod1" && pod.Namespace == testing.TestNamespace {
				foundPod1 = true
				if pod.Status != "Running" {
					t.Errorf("Expected status 'Running', got '%s'", pod.Status)
				}
				if len(pod.Containers) != 1 {
					t.Errorf("Expected 1 container, got %d", len(pod.Containers))
				}
				if pod.Containers[0].Name != "test-container" {
					t.Errorf("Expected container name 'test-container', got '%s'", pod.Containers[0].Name)
				}
			}
			if pod.Name == "pod2" && pod.Namespace == testing.TestNamespace {
				foundPod2 = true
			}
		}
		
		if !foundPod1 {
			t.Error("Expected to find pod1")
		}
		if !foundPod2 {
			t.Error("Expected to find pod2")
		}
	})
	
	t.Run("MultipleContainers", func(t *testing.T) {
		containers := []corev1.Container{
			{Name: "app", Image: "nginx:latest"},
			{Name: "sidecar", Image: "redis:latest"},
		}
		pod := testing.CreateTestPod("multi-container-pod", testing.TestNamespace, containers...)
		
		clientset := testing.NewFakeKubernetesClient(pod)
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		pods, err := client.ListPods(ctx, testing.TestNamespace)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if len(pods) != 1 {
			t.Fatalf("Expected 1 pod, got %d", len(pods))
		}
		
		podInfo := pods[0]
		if len(podInfo.Containers) != 2 {
			t.Errorf("Expected 2 containers, got %d", len(podInfo.Containers))
		}
		
		// Check container readiness
		for _, container := range podInfo.Containers {
			if !container.Ready {
				t.Errorf("Expected container '%s' to be ready", container.Name)
			}
		}
	})
}

func TestGetPod(t *testing.T) {
	ctx := context.Background()
	
	t.Run("ExistingPod", func(t *testing.T) {
		pod := testing.CreateTestPod(testing.TestPod, testing.TestNamespace)
		clientset := testing.NewFakeKubernetesClient(pod)
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		retrievedPod, err := client.GetPod(ctx, testing.TestNamespace, testing.TestPod)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if retrievedPod.Name != testing.TestPod {
			t.Errorf("Expected pod name '%s', got '%s'", testing.TestPod, retrievedPod.Name)
		}
		
		if retrievedPod.Namespace != testing.TestNamespace {
			t.Errorf("Expected namespace '%s', got '%s'", testing.TestNamespace, retrievedPod.Namespace)
		}
	})
	
	t.Run("NonExistentPod", func(t *testing.T) {
		clientset := testing.NewFakeKubernetesClient()
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		_, err := client.GetPod(ctx, testing.TestNamespace, "nonexistent-pod")
		if err == nil {
			t.Error("Expected error for non-existent pod")
		}
	})
}

func TestListNamespaces(t *testing.T) {
	ctx := context.Background()
	
	t.Run("EmptyCluster", func(t *testing.T) {
		clientset := testing.NewFakeKubernetesClient()
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		namespaces, err := client.ListNamespaces(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if len(namespaces) != 0 {
			t.Errorf("Expected 0 namespaces, got %d", len(namespaces))
		}
	})
	
	t.Run("WithNamespaces", func(t *testing.T) {
		ns1 := testing.CreateTestNamespace("namespace1")
		ns2 := testing.CreateTestNamespace("namespace2")
		
		clientset := testing.NewFakeKubernetesClient(ns1, ns2)
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		namespaces, err := client.ListNamespaces(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if len(namespaces) != 2 {
			t.Errorf("Expected 2 namespaces, got %d", len(namespaces))
		}
		
		// Check namespace names
		foundNs1 := false
		foundNs2 := false
		for _, ns := range namespaces {
			if ns == "namespace1" {
				foundNs1 = true
			}
			if ns == "namespace2" {
				foundNs2 = true
			}
		}
		
		if !foundNs1 {
			t.Error("Expected to find namespace1")
		}
		if !foundNs2 {
			t.Error("Expected to find namespace2")
		}
	})
}

func TestCreateEphemeralContainer(t *testing.T) {
	ctx := context.Background()
	
	t.Run("ValidPod", func(t *testing.T) {
		pod := testing.CreateTestPod(testing.TestPod, testing.TestNamespace)
		clientset := testing.NewFakeKubernetesClient(pod)
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		err := client.CreateEphemeralContainer(ctx, testing.TestNamespace, testing.TestPod, testing.TestContainer)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
	
	t.Run("NonExistentPod", func(t *testing.T) {
		clientset := testing.NewFakeKubernetesClient()
		client := &Client{clientset: clientset, config: testing.GetTestConfig()}
		
		err := client.CreateEphemeralContainer(ctx, testing.TestNamespace, "nonexistent", testing.TestContainer)
		if err == nil {
			t.Error("Expected error for non-existent pod")
		}
	})
}

func TestEphemeralToString(t *testing.T) {
	t.Run("BasicEphemeral", func(t *testing.T) {
		ephemeral := corev1.EphemeralContainer{
			EphemeralContainerCommon: corev1.EphemeralContainerCommon{
				Name:  "test-ephemeral",
				Image: "alpine:latest",
				Command: []string{"/bin/sh"},
				Stdin:  true,
				TTY:    true,
			},
			TargetContainerName: "target-container",
		}
		
		result := ephemeralToString(ephemeral)
		
		if result == "" {
			t.Error("Expected non-empty result")
		}
		
		// Check that it contains expected fields
		if !contains(result, `"name":"test-ephemeral"`) {
			t.Error("Expected name field in result")
		}
		if !contains(result, `"image":"alpine:latest"`) {
			t.Error("Expected image field in result")
		}
		if !contains(result, `"stdin":true`) {
			t.Error("Expected stdin field in result")
		}
		if !contains(result, `"tty":true`) {
			t.Error("Expected tty field in result")
		}
		if !contains(result, `"targetContainerName":"target-container"`) {
			t.Error("Expected targetContainerName field in result")
		}
	})
	
	t.Run("WithSecurityContext", func(t *testing.T) {
		privileged := true
		ephemeral := corev1.EphemeralContainer{
			EphemeralContainerCommon: corev1.EphemeralContainerCommon{
				Name:  "test-ephemeral",
				Image: "alpine:latest",
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
			},
		}
		
		result := ephemeralToString(ephemeral)
		
		if !contains(result, `"securityContext":{"privileged":true}`) {
			t.Error("Expected security context with privileged flag")
		}
	})
}

func TestArrayToString(t *testing.T) {
	t.Run("EmptyArray", func(t *testing.T) {
		result := arrayToString([]string{})
		expected := "[]"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("SingleElement", func(t *testing.T) {
		result := arrayToString([]string{"/bin/sh"})
		expected := `["/bin/sh"]`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("MultipleElements", func(t *testing.T) {
		result := arrayToString([]string{"/bin/sh", "-c", "echo hello"})
		expected := `["/bin/sh","-c","echo hello"]`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
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
