package fetch

import (
	"fmt"
	"io"
	"net/http"
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

// HTTP GET from endpoint returns byte array of data for unmarshaling
func (f *FetchData) Get(endpoint, apiKey, qParam string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s%s?key=%s&q=%s", f.Scheme, f.Host, endpoint, apiKey, qParam)
	safeURL := fmt.Sprintf("%s://%s%s?q=%s", f.Scheme, f.Host, endpoint, qParam)
	fmt.Println("Fetching:", safeURL)

	resp, err := http.Get(url)
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
