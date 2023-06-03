package client

import (
	"encoding/json"
	"fmt"
	"log"
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
			err := contracts.HttpError{}
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

func logRequest(client *resty.Client, request *resty.Request) {
	log.Printf("Request URL: %s", request.URL)
	log.Printf("Request Method: %s", request.Method)
	log.Printf("Request Body: %s", request.Body)
}
