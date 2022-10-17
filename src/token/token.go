package token

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

const (
	TypeBot    string = "Bot"
	TypeNormal string = "Bearer"
)

type Token struct {
	AppId   uint64 `yaml:"appid"`
	Token   string `yaml:"token"`
	BotType string `yaml:"botType"`
}

var AccessToken Token

func init() {
	content, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("读取配置文件出错， err = ", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(content, &AccessToken)
	if err != nil {
		log.Println("解析配置文件出错， err = ", err)
		os.Exit(1)
	}
	AccessToken.BotType = TypeBot
	log.Println(AccessToken)
}

func SetToken(appId uint64, token string, botType string) Token {
	AccessToken.AppId = appId
	AccessToken.Token = token
	if botType != "" {
		AccessToken.BotType = botType
	} else {
		AccessToken.BotType = TypeBot
	}
	return AccessToken
}

// GetString 获取授权头字符串
func (t *Token) GetString() string {
	if t.BotType == TypeNormal {
		return t.Token
	}
	return fmt.Sprintf("%v.%s", t.AppId, t.Token)
}