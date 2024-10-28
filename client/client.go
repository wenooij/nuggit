package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/wenooij/nuggit/api"
)

const defaultBackendAddr = "http://localhost:9402"

type Client struct {
	backendURL *url.URL
}

func NewClient() *Client {
	u, _ := url.Parse(defaultBackendAddr)
	return &Client{backendURL: u}
}

func (c *Client) apiURL(path string) (string, error) {
	u, err := c.backendURL.Parse(path)
	if err != nil {
		return "", err
	}
	return u.String(), err
}

func (c *Client) marshalRequest(a any) (io.Reader, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (c *Client) newRequest(method, path string, payload any) (*http.Request, error) {
	u, err := c.apiURL(path)
	if err != nil {
		return nil, err
	}
	body, err := c.marshalRequest(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) handleResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp.Status, resp.Body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := json.MarshalIndent(json.RawMessage(body), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data)) // TODO: Any other actions here?
	return nil
}

func (c *Client) handleError(status string, body io.ReadCloser) error {
	errMessage, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	defer body.Close()

	var m map[string]string
	if err := json.Unmarshal(errMessage, &m); err != nil {
		return err
	}
	return fmt.Errorf("%s (%s)", m["reason"], status)
}

func (c *Client) DisablePipe(name, digest string) error {
	req, err := c.newRequest("POST", "/api/pipes/disable", api.DisablePipeRequest{
		Name:   name,
		Digest: digest,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return c.handleResponse(resp)
}

func (c *Client) EnablePipe(name, digest string) error {
	req, err := c.newRequest("POST", "/api/pipes/enable", api.EnablePipeRequest{
		Name:   name,
		Digest: digest,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return c.handleResponse(resp)
}
