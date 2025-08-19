package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// FetchData is a generic API client for any service
type FetchData struct {
	Scheme string
	Host   string
	Port   int
}

// NewFetchData is the constructor for a generic API client
func NewFetchData(scheme, host string, port int) *FetchData {
	return &FetchData{
		Scheme: scheme,
		Host:   host,
		Port:   port,
	}
}

type Path struct {
	Endpoint string
	Params   map[string]string
	Headers  map[string]string
}

// HTTP GET from endpoint returns byte array of data for unmarshaling
func (f *FetchData) Get(path Path) ([]byte, error) {
	url := url.URL{
		Scheme: f.Scheme,
		Host:   f.Host,
		Path:   path.Endpoint,
	}

	q := url.Query()
	for k, v := range path.Params {
		q.Set(k, v)
	}
	url.RawQuery = q.Encode()

	fmt.Println("Fetching:", url.String())

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
