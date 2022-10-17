package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	wss "github.com/gorilla/websocket"
	"log"
	"robot/src/dto"
	"robot/src/handler"
	"time"
)

type messageChan chan *dto.WSPayload
type closeErrorChan chan error

// Client websocket 连接客户端
type WebsocketClient struct {
	version         int
	conn            *wss.Conn
	messageQueue    messageChan
	session         *dto.Session
	user            *dto.WSUser
	closeChan       closeErrorChan
	heartBeatTicker *time.Ticker // 用于维持定时心跳
}

// DefaultQueueSize 监听队列的缓冲长度
const DefaultQueueSize = 10000

// New 新建一个连接对象
func (wsc *WebsocketClient) New(session dto.Session) *WebsocketClient {
	return &WebsocketClient{
		messageQueue:    make(messageChan, DefaultQueueSize),
		session:         &session,
		closeChan:       make(closeErrorChan, 10),
		heartBeatTicker: time.NewTicker(60 * time.Second), // 先给一个默认 ticker，在收到 hello 包之后，会 reset
	}
}

// Connect 连接到 websocket
func (wsc *WebsocketClient) Connect() error {
	if wsc.session.URL == "" {
		return errors.New("session.URL error")
	}

	var err error
	wsc.conn, _, err = wss.DefaultDialer.Dial(wsc.session.URL, nil)
	if err != nil {
		log.Printf("%s, connect err: %v", wsc.session, err)
		return err
	}
	log.Printf("%s, url %s, connected", wsc.session, wsc.session.URL)

	return nil
}

// Listening 开始监听，会阻塞进程，内部会从事件队列不断的读取事件，解析后投递到注册的 event handler，如果读取消息过程中发生错误，会循环
// 定时心跳也在这里维护
func (wsc *WebsocketClient) Listening() error {
	defer wsc.Close()
	// reading message
	go wsc.readMessageToQueue()
	// read message from queue and handle,in goroutine to avoid business logic block closeChan and heartBeatTicker
	go wsc.listenMessageAndHandle()


	// handler message
	for {
		select {
		case err := <-wsc.closeChan:
			return err
		case <-wsc.heartBeatTicker.C:
			log.Printf("%s listened heartBeat", wsc.session)
			heartBeatEvent := &dto.WSPayload{
				WSPayloadBase: dto.WSPayloadBase{
					OPCode: dto.WSHeartbeat,
				},
				Data: wsc.session.LastSeq,
			}
			// 不处理错误，Write 内部会处理，如果发生发包异常，会通知主协程退出
			_ = wsc.Write(heartBeatEvent)
		}
	}
}

// Write 往 ws 写入数据
func (wsc *WebsocketClient) Write(message *dto.WSPayload) error {
	m, _ := json.Marshal(message)
	log.Printf("%s write %s message, %v", wsc.session, dto.OPMeans(message.OPCode), string(m))

	if err := wsc.conn.WriteMessage(wss.TextMessage, m); err != nil {
		log.Printf("%s WriteMessage failed, %v", wsc.session, err)
		wsc.closeChan <- err
		return err
	}
	return nil
}

// Resume 重连
func (wsc *WebsocketClient) Resume() error {
	payload := &dto.WSPayload{
		Data: &dto.WSResumeData{
			Token:     wsc.session.Token.GetString(),
			SessionID: wsc.session.ID,
			Seq:       wsc.session.LastSeq,
		},
	}
	payload.OPCode = dto.WSResume // 内嵌结构体字段，单独赋值
	return wsc.Write(payload)
}

// Identify 对一个连接进行鉴权，并声明监听的 shard 信息
func (wsc *WebsocketClient) Identify() error {
	// 避免传错 intent
	if wsc.session.Intent == 0 {
		wsc.session.Intent = dto.IntentGuilds
	}
	payload := &dto.WSPayload{
		Data: &dto.WSIdentityData{
			Token:   wsc.session.Token.GetString(),
			Intents: wsc.session.Intent,
			Shard: []uint32{
				wsc.session.Shards.ShardID,
				wsc.session.Shards.ShardCount,
			},
		},
	}
	payload.OPCode = dto.WSIdentity
	return wsc.Write(payload)
}

// Close 关闭连接
func (wsc *WebsocketClient) Close() {
	if err := wsc.conn.Close(); err != nil {
		log.Printf("%s, close conn err: %v", wsc.session, err)
	}
	wsc.heartBeatTicker.Stop()
}

// Session 获取client的session信息
func (wsc *WebsocketClient) Session() *dto.Session {
	return wsc.session
}

func (wsc *WebsocketClient) readMessageToQueue() {
	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			log.Printf("%s read message failed, %v, message %s", wsc.session, err, string(message))
			close(wsc.messageQueue)
			wsc.closeChan <- err
			return
		}
		payload := &dto.WSPayload{}
		if err := json.Unmarshal(message, payload); err != nil {
			log.Printf("%s json failed, %v", wsc.session, err)
			continue
		}
		payload.RawMessage = message
		log.Printf("%s receive %s message, %s", wsc.session, dto.OPMeans(payload.OPCode), string(message))
		// 处理内置的一些事件，如果处理成功，则这个事件不再投递给业务
		if wsc.isHandleBuildIn(payload) {
			continue
		}
		wsc.messageQueue <- payload
	}
}

func (wsc *WebsocketClient) listenMessageAndHandle() {
	defer func() {
		// panic，一般是由于业务自己实现的 handle 不完善导致
		// 打印日志后，关闭这个连接，进入重连流程
		if err := recover(); err != nil {
			log.Printf("%v,%s", err, wsc.session)
			wsc.closeChan <- fmt.Errorf("panic: %v", err)
		}
	}()
	for payload := range wsc.messageQueue {
		wsc.saveSeq(payload.Seq)
		// ready 事件需要特殊处理
		if payload.Type == "READY" {
			wsc.readyHandler(payload)
			continue
		}
		// 解析具体事件，并投递给业务注册的 handler
		if err := handler.ParseAndHandle(payload); err != nil {
			log.Printf("%s parseAndHandle failed, %v", wsc.session, err)
		}
	}
	log.Printf("%s message queue is closed", wsc.session)
}

func (wsc *WebsocketClient) saveSeq(seq uint32) {
	if seq > 0 {
		wsc.session.LastSeq = seq
	}
}

// isHandleBuildIn 内置的事件处理，处理那些不需要业务方处理的事件
// return true 的时候说明事件已经被处理了
func (wsc *WebsocketClient) isHandleBuildIn(payload *dto.WSPayload) bool {
	switch payload.OPCode {
	case dto.WSHello: // 接收到 hello 后需要开始发心跳
		wsc.startHeartBeatTicker(payload.RawMessage)
	case dto.WSHeartbeatAck: // 心跳 ack 不需要业务处理
	case dto.WSReconnect: // 达到连接时长，需要重新连接，此时可以通过 resume 续传原连接上的事件
	case dto.WSInvalidSession: // 无效的 sessionLog，需要重新鉴权
	default:
		return false
	}
	return true
}

// startHeartBeatTicker 启动定时心跳
func (wsc *WebsocketClient) startHeartBeatTicker(message []byte) {
	helloData := &dto.WSHelloData{}
	if err := handler.ParseData(message, helloData); err != nil {
		log.Printf("%s hello data parse failed, %v, message %v", wsc.session, err, message)
	}
	// 根据 hello 的回包，重新设置心跳的定时器时间
	wsc.heartBeatTicker.Reset(time.Duration(helloData.HeartbeatInterval) * time.Millisecond)
}

// readyHandler 针对ready返回的处理，需要记录 sessionID 等相关信息
func (wsc *WebsocketClient) readyHandler(payload *dto.WSPayload) {
	readyData := &dto.WSReadyData{}
	if err := handler.ParseData(payload.RawMessage, readyData); err != nil {
		log.Printf("%s parseReadyData failed, %v, message %v", wsc.session, err, payload.RawMessage)
	}
	wsc.version = readyData.Version
	// 基于 ready 事件，更新 session 信息
	wsc.session.ID = readyData.SessionID
	wsc.session.Shards.ShardID = readyData.Shard[0]
	wsc.session.Shards.ShardCount = readyData.Shard[1]
	wsc.user = &dto.WSUser{
		ID:       readyData.User.ID,
		Username: readyData.User.Username,
		Bot:      readyData.User.Bot,
	}
}
