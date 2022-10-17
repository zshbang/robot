package application

import (
	"context"
	"fmt"
	"log"
	"os"
	"robot/src/client"
	"robot/src/handler"
	"robot/src/token"
	"robot/src/websocket"
)

var Ctx context.Context
func SetStartContext(ctx context.Context)  {
	Ctx = ctx
}

func Run()  {
	ws, err := client.ClientImpl.GetWebsocketAccessPoint(Ctx)
	if err != nil {
		log.Fatalln("websocket错误， err = ", err)
		os.Exit(1)
	}
	fmt.Println(ws)
	_ = websocket.Start(ws, &token.AccessToken, &handler.Intent)
}
