package openapi

import (
	"context"
	"robot/src/dto"
	"time"
)

type OpenApi interface {
	Base
	WebsocketApi
	MessageApi
}

type Base interface {
	WithTimeout(duration time.Duration) OpenApi

	Transport(ctx context.Context, method, url string, body interface{}) ([]byte, error)

	GetToken() string
}

type WebsocketApi interface {
	GetWebsocketAccessPoint(ctx context.Context) (*dto.WebsocketAP, error)
}

type MessageApi interface {
	PostMessage(ctx context.Context, channelID string, msg *dto.MessageToCreate) (*dto.Message, error)
}
