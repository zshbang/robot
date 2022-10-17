package main

import (
	"context"
	"robot/src/application"
	"robot/src/client"
	"robot/src/dto"
	"robot/src/handler"
	"robot/src/main/function/calculator"
	"robot/src/main/function/weather"
	"strings"
)

var api = client.ClientImpl
var ctx context.Context

//var channelId = "" //保存子频道的id

func main() {
	ctx = context.Background()
	application.SetStartContext(ctx)
	handler.RegisterHandlers(atMessageEventHandler)
	application.Run()

}

func atMessageEventHandler(event *dto.WSPayload, data *dto.WSATMessageData) error {
	//channelId = data.ChannelID //当@机器人时，保存ChannelId，主动消息需要 channelId 才能发送出去
	if strings.HasSuffix(data.Content, "> hello") {
		//获取深圳的天气数据
		weatherData := weather.GetWeatherByCity("深圳")
		_, _ = api.PostMessage(ctx, data.ChannelID, &dto.MessageToCreate{MsgID: data.ID,
			Content: weatherData.ResultData.CityNm + " " + weatherData.ResultData.Weather + " " + weatherData.ResultData.Days + " " + weatherData.ResultData.Week,
			Image:   weatherData.ResultData.WeatherIcon, //天气图片
		})
	}

	if strings.HasSuffix(data.Content, "=") {
		index := strings.Index(data.Content, ">")
		expression := data.Content[index+1:]
		result := calculator.HandleAndCalculate(expression)

		_, _ = api.PostMessage(ctx, data.ChannelID, &dto.MessageToCreate{MsgID: data.ID,
			Content: "结果为：" + result,
		})
	}
	return nil
}
