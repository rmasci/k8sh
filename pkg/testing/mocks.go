package testing

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// MockExecutor implements remotecommand.Executor for testing
type MockExecutor struct {
	Output string
	Error  error
}

func (m *MockExecutor) Stream(ctx context.Context, options remotecommand.StreamOptions) error {
	if m.Error != nil {
		return m.Error
	}
	
	if options.Stdout != nil {
		fmt.Fprint(options.Stdout, m.Output)
	}
	
	return nil
}

func (m *MockExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	return m.Stream(ctx, options)
}

// MockSPDYExecutor creates a mock SPDY executor
type MockSPDYExecutor struct {
	Executors map[string]*MockExecutor
}

func NewMockSPDYExecutor() *MockSPDYExecutor {
	return &MockSPDYExecutor{
		Executors: make(map[string]*MockExecutor),
	}
}

func (m *MockSPDYExecutor) AddExecutor(key string, output string, err error) {
	m.Executors[key] = &MockExecutor{Output: output, Error: err}
}

// MockConfig implements rest.Config for testing
type MockConfig struct {
	Host string
}

// NewFakeKubernetesClient creates a fake Kubernetes client with test data
func NewFakeKubernetesClient(objects ...runtime.Object) *fake.Clientset {
	return fake.NewClientset(objects...)
}

// CreateTestPod creates a test pod for testing
func CreateTestPod(name, namespace string, containers ...corev1.Container) *corev1.Pod {
	if len(containers) == 0 {
		containers = []corev1.Container{
			{
				Name:  "test-container",
				Image: "nginx:latest",
			},
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  containers[0].Name,
					Ready: true,
				},
			},
		},
	}

	return pod
}

// CreateTestNamespace creates a test namespace
func CreateTestNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// MockRESTClient provides a mock REST client interface
type MockRESTClient struct {
	PostFunc func() *MockRESTRequest
}

func (m *MockRESTClient) Post() *MockRESTRequest {
	if m.PostFunc != nil {
		return m.PostFunc()
	}
	return &MockRESTRequest{}
}

// MockRESTRequest provides a mock REST request
type MockRESTRequest struct {
	ResourceFunc    func(string) *MockRESTRequest
	NamespaceFunc   func(string) *MockRESTRequest
	NameFunc        func(string) *MockRESTRequest
	SubResourceFunc func(string) *MockRESTRequest
	VersionedParamsFunc func(interface{}, interface{}) *MockRESTRequest
	URLFunc         func() string
	DoFunc         func() error
}

func (m *MockRESTRequest) Resource(resource string) *MockRESTRequest {
	if m.ResourceFunc != nil {
		return m.ResourceFunc(resource)
	}
	return m
}

func (m *MockRESTRequest) Namespace(namespace string) *MockRESTRequest {
	if m.NamespaceFunc != nil {
		return m.NamespaceFunc(namespace)
	}
	return m
}

func (m *MockRESTRequest) Name(name string) *MockRESTRequest {
	if m.NameFunc != nil {
		return m.NameFunc(name)
	}
	return m
}

func (m *MockRESTRequest) SubResource(subResource string) *MockRESTRequest {
	if m.SubResourceFunc != nil {
		return m.SubResourceFunc(subResource)
	}
	return m
}

func (m *MockRESTRequest) VersionedParams(params interface{}, codec interface{}) *MockRESTRequest {
	if m.VersionedParamsFunc != nil {
		return m.VersionedParamsFunc(params, codec)
	}
	return m
}

func (m *MockRESTRequest) URL() string {
	if m.URLFunc != nil {
		return m.URLFunc()
	}
	return "http://mock-url"
}

func (m *MockRESTRequest) Do() error {
	if m.DoFunc != nil {
		return m.DoFunc()
	}
	return nil
}

// Test utilities
const (
	TestNamespace = "test-namespace"
	TestPod       = "test-pod"
	TestContainer = "test-container"
	TestFilePath  = "/test/file.txt"
	TestDirPath   = "/test/dir"
)

// GetTestConfig returns a mock REST config for testing
func GetTestConfig() *rest.Config {
	return &rest.Config{
		Host: "http://localhost:8080",
	}
}
