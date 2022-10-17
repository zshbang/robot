package websocket

import (
	"log"
	"math"
	"robot/src/dto"
	"robot/src/token"
	"time"
)

var sessionChan chan dto.Session
var wsc *WebsocketClient



func Start(apInfo *dto.WebsocketAP, token *token.Token, intents *dto.Intent) error  {
	sessionChan = make(chan dto.Session, apInfo.Shards)
	for i := uint32(0); i < apInfo.Shards; i++ {
		session := dto.Session{
			URL:     apInfo.URL,
			Token:   *token,
			Intent:  *intents,
			LastSeq: 0,
			Shards: dto.ShardConfig{
				ShardID:    i,
				ShardCount: apInfo.Shards,
			},
		}
		sessionChan <- session
	}
	startInterval := CalcInterval(apInfo.SessionStartLimit.MaxConcurrency)
	for session := range sessionChan {
		// MaxConcurrency 代表的是每 5s 可以连多少个请求
		time.Sleep(startInterval)
		go newConnect(session)
	}
	return nil
}

func newConnect(session dto.Session) {
	defer func() {
		// panic 留下日志，放回 session
		if err := recover(); err != nil {
			log.Println(err, &session)
			sessionChan <- session
		}
	}()

	wsc = wsc.New(session)
	if err := wsc.Connect(); err != nil {
		log.Println(err)
		sessionChan <- session // 连接失败，丢回去队列排队重连
		return
	}
	var err error
	// 如果 session id 不为空，则执行的是 resume 操作，如果为空，则执行的是 identify 操作
	if session.ID != "" {
		err = wsc.Resume()
	} else {
		// 初次鉴权
		err = wsc.Identify()
	}
	if err != nil {
		log.Println("[ws/session] Identify/Resume err " + err.Error())
		return
	}
	if err := wsc.Listening(); err != nil {
		log.Println("[ws/session] Listening err " + err.Error())
		sessionChan <- *wsc.Session()
		return
	}
}

// concurrencyTimeWindowSec 并发时间窗口，单位秒
const concurrencyTimeWindowSec = 2

// CalcInterval 根据并发要求，计算连接启动间隔
func CalcInterval(maxConcurrency uint32) time.Duration {
	if maxConcurrency == 0 {
		maxConcurrency = 1
	}
	f := math.Round(concurrencyTimeWindowSec / float64(maxConcurrency))
	if f == 0 {
		f = 1
	}
	return time.Duration(f) * time.Second
}