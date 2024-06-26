/*
Boilerplate
*/
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "crds-controller/pkg/apis/sample/v1alpha1"
	"crds-controller/pkg/generated/clientset/versioned/scheme"
	"net/http"

	rest "k8s.io/client-go/rest"
)

type SampleV1alpha1Interface interface {
	RESTClient() rest.Interface
	FoosGetter
}

// SampleV1alpha1Client is used to interact with features provided by the sample.dev group.
type SampleV1alpha1Client struct {
	restClient rest.Interface
}

func (c *SampleV1alpha1Client) Foos(namespace string) FooInterface {
	return newFoos(c, namespace)
}

// NewForConfig creates a new SampleV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*SampleV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new SampleV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*SampleV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &SampleV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new SampleV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *SampleV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new SampleV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *SampleV1alpha1Client {
	return &SampleV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *SampleV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
