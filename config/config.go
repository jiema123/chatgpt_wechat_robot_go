package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/qingconglaixueit/wechatbot/pkg/logger"
)

// Configuration 项目配置
type Configuration struct {
	// gpt apikey
	ApiKey string `json:"api_key"`
	// 自动通过好友
	AutoPass bool `json:"auto_pass"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout"`
	// GPT请求最大字符数
	MaxTokens uint `json:"max_tokens"`
	// GPT模型
	Model string `json:"model"`
	// 热度
	Temperature float64 `json:"temperature"`
	// 回复前缀
	ReplyPrefix string `json:"reply_prefix"`
	// 清空会话口令
	SessionClearToken string `json:"session_clear_token"`
	// 规则 system user assistant
	Role string `json:"role"`
	// 画图画口令
	ImageStartKey string `json:"image_start_key"`
	//0-10 的数字
	ImageN int `json:"image_n"`
	//256x256, 512x512, or 1024x1024.
	ImageSize string `json:"image_size"`
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 给配置赋默认值
		config = &Configuration{
			AutoPass:          false,
			SessionTimeout:    60,
			MaxTokens:         512,
			Model:             "gpt-3.5-turbo-0301",
			Temperature:       0.9,
			Role:              "assistant",
			SessionClearToken: "下个问题",
			ImageStartKey:     "画图画",
			ImageN:            10,
			ImageSize:         "256x256",
		}

		// 判断配置文件是否存在，存在直接JSON读取
		_, err := os.Stat("config.json")
		if err == nil {
			f, err := os.Open("config.json")
			if err != nil {
				logger.Danger(fmt.Sprintf("open config error: %v", err))
				return
			}
			defer f.Close()
			encoder := json.NewDecoder(f)
			err = encoder.Decode(config)
			if err != nil {
				logger.Danger(fmt.Sprintf("decode config error: %v", err))
				return
			}
		}
		// 有环境变量使用环境变量
		ApiKey := os.Getenv("APIKEY")
		AutoPass := os.Getenv("AUTO_PASS")
		SessionTimeout := os.Getenv("SESSION_TIMEOUT")
		Model := os.Getenv("MODEL")
		MaxTokens := os.Getenv("MAX_TOKENS")
		Temperature := os.Getenv("TEMPREATURE")
		ReplyPrefix := os.Getenv("REPLY_PREFIX")
		SessionClearToken := os.Getenv("SESSION_CLEAR_TOKEN")
		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if AutoPass == "true" {
			config.AutoPass = true
		}
		if SessionTimeout != "" {
			duration, err := time.ParseDuration(SessionTimeout)
			if err != nil {
				logger.Danger(fmt.Sprintf("config session timeout error: %v, get is %v", err, SessionTimeout))
				return
			}
			config.SessionTimeout = duration
		}
		if Model != "" {
			config.Model = Model
		}
		if MaxTokens != "" {
			max, err := strconv.Atoi(MaxTokens)
			if err != nil {
				logger.Danger(fmt.Sprintf("config max tokens error: %v ,get is %v", err, MaxTokens))
				return
			}
			config.MaxTokens = uint(max)
		}
		if Temperature != "" {
			temp, err := strconv.ParseFloat(Temperature, 64)
			if err != nil {
				logger.Danger(fmt.Sprintf("config temperature error: %v, get is %v", err, Temperature))
				return
			}
			config.Temperature = temp
		}
		if ReplyPrefix != "" {
			config.ReplyPrefix = ReplyPrefix
		}
		if SessionClearToken != "" {
			config.SessionClearToken = SessionClearToken
		}

	})
	if config.ApiKey == "" {
		logger.Danger("config error: api key required")
	}

	return config
}
