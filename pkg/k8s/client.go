package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
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
