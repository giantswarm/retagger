package quay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const base = "https://quay.io"

type Client struct {
	Base                string
	client              *http.Client
	authorizationHeader string
}

func DefaultClient() Client {

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	return Client{
		Base:   base,
		client: netClient,
	}
}

func (c *Client) AuthorizationHeader(ah string) {
	c.authorizationHeader = ah
}

func (c *Client) do(method string, url string, body io.Reader) (int, string, []byte, error) {
	req, err := http.NewRequest(method, c.Base+url, body)
	req.Header.Add("Authorization", c.authorizationHeader)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return -1, "", nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", nil, err
	}
	return resp.StatusCode, resp.Status, b, nil
}

func mustReader(i interface{}) io.Reader {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(b)
}

func getAPIError(status string, body []byte) error {
	return fmt.Errorf("API error: %s- %s", status, string(body))
}
