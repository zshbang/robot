package handler

import (
	"robot/src/dto"
)

// DefaultHandlers 默认的 handler 结构，管理所有支持的 handler 类型
var DefaultHandlers struct {
	ATMessage ATMessageEventHandler
}

// ATMessageEventHandler at 机器人消息事件 handler
type ATMessageEventHandler = func(event *dto.WSPayload, data *dto.WSATMessageData) error
var Intent dto.Intent

func RegisterHandlers(handler interface{}) dto.Intent {
	switch h := handler.(type) {
		case ATMessageEventHandler:
			DefaultHandlers.ATMessage = h
			Intent = 1 << 30
	}
	return Intent
}
