package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// FetchData is a generic API client for any service
type Fetch struct {
	Scheme string
	Host   string
	Port   int
}

// NewFetchData is the constructor for a generic API client
func NewFetch(scheme, host string, port int) *Fetch {
	return &Fetch{
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
func (f *Fetch) Get(path Path) ([]byte, error) {
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

func (f *Fetch) GetJSONHeader(path Path) ([]byte, error) {
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

	fmt.Println("Fetching with JSON header:", url.String())

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.sdmx.structure+json")

	client := &http.Client{}
	resp, err := client.Do(req)
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
