package downloader

import (
	"io"
	"net/http"
)

type IHttpClient interface {
	GenerateRequest(method string, url string, body io.Reader) (*http.Request, error)
	Do(req *http.Request, headers ...*http.Header) (*http.Response, error)
}

type HttpClient struct {
	*http.Client
}

func NewHttpClient(client *http.Client) IHttpClient {
	return &HttpClient{Client: client}
}

func (hc *HttpClient) GenerateRequest(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (hc *HttpClient) Do(req *http.Request, headers ...*http.Header) (*http.Response, error) {
	if len(headers) > 0 && headers[0] != nil {
		for key, header := range headers[0].Clone() {
			req.Header.Set(key, header[0])
		}
	}
	return hc.Do(req)
}
