package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boichique/movie-reviews/contracts"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client  *resty.Client
	baseURL string
}

func New(url string) *Client {
	hc := &http.Client{}
	rc := resty.NewWithClient(hc)
	rc.OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		if response.IsError() {
			err := contracts.HTTPError{}
			_ = json.Unmarshal(response.Body(), &err)

			return &Error{
				Code:    response.StatusCode(),
				Message: err.Message,
			}
		}

		return nil
	})

	return &Client{
		client:  rc,
		baseURL: url,
	}
}

func (c *Client) path(f string, args ...any) string {
	return fmt.Sprintf(c.baseURL+f, args...)
}
