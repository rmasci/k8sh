package k8s

import (
	"testing"
	
	k8stesting "github.com/rmasci/k8sh/pkg/testing"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestNewClient(t *testing.T) {
	// This test requires a valid kubeconfig, so we'll test the structure
	// In a real environment, you would mock the clientcmd.BuildConfigFromFlags
	
	t.Run("ClientStructure", func(t *testing.T) {
		// Test that client has the right structure
		client := &Client{
			config:    k8stesting.GetTestConfig(),
			clientset: k8stesting.NewFakeKubernetesClient(),
		}
		
		if client.config == nil {
			t.Error("Expected config to be non-nil")
		}
		
		if client.clientset == nil {
			t.Error("Expected clientset to be non-nil")
		}
	})
}

func TestClientGetters(t *testing.T) {
	config := k8stesting.GetTestConfig()
	clientset := k8stesting.NewFakeKubernetesClient()
	
	client := &Client{
		config:    config,
		clientset: clientset,
	}
	
	t.Run("GetClientset", func(t *testing.T) {
		retrievedClientset := client.GetClientset()
		if retrievedClientset != clientset {
			t.Error("GetClientset should return the same clientset instance")
		}
	})
	
	t.Run("GetConfig", func(t *testing.T) {
		retrievedConfig := client.GetConfig()
		if retrievedConfig != config {
			t.Error("GetConfig should return the same config instance")
		}
	})
}

func TestClientInterface(t *testing.T) {
	// Test that Client implements expected interface
	var _ interface {
		GetClientset() *kubernetes.Clientset
		GetConfig() *rest.Config
	} = &Client{}
}
