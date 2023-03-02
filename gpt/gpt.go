package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/qingconglaixueit/wechatbot/config"
)

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
	Error   struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

type MessageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChoiceItem struct {
	Messages MessageItem `json:"message"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model    string        `json:"model"`
	Messages []MessageItem `json:"messages"`
}

//curl --location --request POST 'https://api.openai.com/v1/chat/completions' \
//--header 'Authorization: Bearer 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' \
//--header 'User-Agent: apifox/1.0.0 (https://www.apifox.cn)' \
//--header 'Content-Type: application/json' \
//--data-raw '{
//"model": "gpt-3.5-turbo-0301",
//"messages": [
//{
//"role": "assistant",
//"content": "who are you"
//}
//]
//}'
func Completions(msg string) (string, error) {
	var gptResponseBody *ChatGPTResponseBody
	var resErr error
	var messages []MessageItem
	cfg := config.LoadConfig()
	for retry := 1; retry <= 3; retry++ {
		if retry > 1 {
			time.Sleep(time.Duration(retry-1) * 100 * time.Millisecond)
		}
		msgItem := MessageItem{
			Role:    cfg.Role,
			Content: msg,
		}
		messages = append(messages, msgItem)
		gptResponseBody, resErr = httpRequestCompletions(messages, retry)
		if resErr != nil {
			log.Printf("gpt request(%d) error: %v\n", retry, resErr)
			continue
		}
		if gptResponseBody.Error.Message == "" {
			break
		}
	}
	if resErr != nil {
		return "", resErr
	}
	var reply string
	if gptResponseBody != nil && len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Messages.Content
		log.Printf("gpt reply string %v\n", reply)
	}
	return reply, nil
}

func httpRequestCompletions(msg []MessageItem, runtimes int) (*ChatGPTResponseBody, error) {
	cfg := config.LoadConfig()
	if cfg.ApiKey == "" {
		return nil, errors.New("api key required")
	}

	requestBody := ChatGPTRequestBody{
		Model:    cfg.Model,
		Messages: msg,
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal requestBody error: %v", err)
	}

	log.Printf("gpt request(%d) json: %s\n", runtimes, string(requestData))

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiKey)
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do error: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll error: %v", err)
	}

	log.Printf("gpt response(%d) json: %s\n", runtimes, string(body))

	gptResponseBody := &ChatGPTResponseBody{}
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal responseBody error: %v", err)
	}
	return gptResponseBody, nil
}
