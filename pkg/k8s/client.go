package k8s

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var initClientError error
var client *Client
var once sync.Once

type Client struct {
	CoreV1    *kubernetes.Clientset
	Config    *rest.Config
	Namespace string
}

type PortForward struct {
	ReadyChannel   chan struct{}
	StopChannel    chan struct{}
	Name           string
	Namespace      string
	ForwardedPorts []portforward.ForwardedPort
}

func (o *Client) CreateDialer(name string, namespace string) (*httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(o.Config)
	if err != nil {
		return nil, err
	}

	req := o.CoreV1.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(name).
		SubResource("portforward")

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, req.URL())

	return &dialer, nil
}

func (o *Client) ForwardPorts(name string, namespace string, ports []string) (PortForward, error) {
	dialer, err := o.CreateDialer(name, namespace)
	if err != nil {
		return PortForward{}, err
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(*dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return PortForward{}, err
	}

	go func() {
		err = forwarder.ForwardPorts()
	}()

	select {
	case <-readyChan:
	case <-stopChan:
		if err != nil {
			return PortForward{}, err
		}
	}

	forwardedPorts, err := forwarder.GetPorts()
	if err != nil {
		return PortForward{}, nil
	}

	return PortForward{
		StopChannel:    stopChan,
		ReadyChannel:   readyChan,
		Name:           name,
		Namespace:      namespace,
		ForwardedPorts: forwardedPorts,
	}, nil
}

func loadClientConfig(kubeConfigPath string) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	if kubeConfigPath != "" {
		loadingRules.ExplicitPath = kubeConfigPath
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
}

func createClient(kubeConfigPath string) (*Client, error) {
	clientConfig := loadClientConfig(kubeConfigPath)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	coreV1, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, err
	}

	return &Client{
		CoreV1: coreV1, Config: restConfig,
		Namespace: namespace,
	}, nil
}

func InitClient(kubeConfigPath string) (*Client, error) {
	once.Do(func() {
		client, initClientError = createClient(kubeConfigPath)
	})
	return client, initClientError
}

func GetClient() *Client {
	if client == nil {
		panic(fmt.Errorf("Client not initialized"))
	}
	return client
}
