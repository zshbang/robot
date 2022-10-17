package client

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"robot/src/dto"
	"robot/src/resourse"
	"robot/src/token"
	"time"
)

type Client struct {
	token       *token.Token
	timeout     time.Duration
	restyClient *resty.Client
}


func init() {
	createApiInstance(&token.AccessToken)
}

// MaxIdleConns 默认指定空闲连接池大小
const MaxIdleConns = 3000

var ClientImpl = &Client{}

func createApiInstance(token *token.Token)  {
	ClientImpl.token = token
	ClientImpl.timeout = 3 * time.Second
	ClientImpl.setRestyClient()
}

func (c *Client) setRestyClient() {
	c.restyClient = resty.New().
		SetTransport(createTransport(nil, MaxIdleConns)). // 自定义 transport
		SetTimeout(c.timeout).
		SetAuthToken(c.token.GetString()).
		SetAuthScheme(c.token.BotType).
		SetHeader("User-Agent", "v0.0.1" + "BotGoSDK")

}

func createTransport(localAddr net.Addr, idleConns int) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          idleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   idleConns,
		MaxConnsPerHost:       idleConns,
	}

}
func (c *Client) request(ctx context.Context) *resty.Request {
	return c.restyClient.R().SetContext(ctx)
}

// getURL 获取接口地址，会处理沙箱环境判断
func (c *Client) getURL(endpoint string) string {
	d := resourse.Domain
	return fmt.Sprintf("%s://%s%s", "https", d, endpoint)
}

// PostMessage 发消息
func (c *Client) PostMessage(ctx context.Context, channelID string, msg *dto.MessageToCreate) (*dto.Message, error) {
	resp, err := c.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetBody(msg).
		Post(c.getURL(resourse.MessagesURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}