package url

import "resty.dev/v3"

var baseURL = "https://api.myanimelist.net/v2/anime"

const ClientIDHeader = "X-MAL-CLIENT-ID"

type Client struct {
	client *resty.Client
}

func NewClient() *Client {
	c := resty.New()
	return &Client{client: c}
}
