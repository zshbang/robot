package client

import (
	"context"
	"robot/src/openapi"
	"time"
)

func (c *Client) WithTimeout(duration time.Duration) openapi.OpenApi {
	c.restyClient.SetTimeout(duration)
	return c
}

func (c *Client) Transport(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	resp, err := c.request(ctx).SetBody(body).Execute(method, url)
	return resp.Body(), err
}

func (c *Client) GetToken() string {
	return c.token.Token
}