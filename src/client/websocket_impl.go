package client

import (
	"context"
	"robot/src/dto"
	"robot/src/resourse"
)

func (c *Client) GetWebsocketAccessPoint(ctx context.Context) (*dto.WebsocketAP, error) {
	resp, err := c.request(ctx).
		SetResult(dto.WebsocketAP{}).
		Get(c.getURL(resourse.GatewayBotURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.WebsocketAP), nil
}

