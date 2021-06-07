package downloader

import "net/http"

type IHttpClient interface {
	GenerateRequest(url string) (*http.Request, error)
	Head(url string) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (resp *http.Response, err error)
}

type HttpClient struct {
	*http.Client
}

func NewHttpClient(client *http.Client) IHttpClient {
	return &HttpClient{Client: client}
}

func (c *HttpClient) GenerateRequest(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, url, nil)
}
