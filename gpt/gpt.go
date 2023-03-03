package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qingconglaixueit/wechatbot/config"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
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

type ImageGPTRequestBody struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type ImageGPTResponseBody struct {
	Created int            `json:"created"`
	Data    []ImageUrlItem `json:"data"`
	Error   struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

type ImageUrlItem struct {
	Url string `json:"url"`
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
	cfg := config.LoadConfig()

	// 对msg 内容进行判断 进去画画api
	if strings.Contains(msg, cfg.ImageStartKey) {
		return imageProcess(strings.Replace(msg, cfg.ImageStartKey, "", 1))
	}

	return textProcess(msg)

}

func imageProcess(msg string) (string, error) {
	var gptResponseBody *ImageGPTResponseBody
	var resErr error
	var reply string

	for retry := 1; retry <= 3; retry++ {
		if retry > 1 {
			time.Sleep(time.Duration(retry-1) * 100 * time.Millisecond)
		}
		gptResponseBody, resErr = httpRequestImagesGenerations(msg, retry)
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

	if gptResponseBody != nil && len(gptResponseBody.Data) > 0 {
		for i := 0; i < len(gptResponseBody.Data); i++ {
			num := i + 1
			reply += "第(" + fmt.Sprintf("%d", num) + ")幅图画\n" + gptResponseBody.Data[i].Url + "\n"
		}
		log.Printf("gpt reply string %v\n", reply)
	}
	return reply, nil
}

func textProcess(msg string) (string, error) {
	cfg := config.LoadConfig()
	var gptResponseBody *ChatGPTResponseBody
	var resErr error
	var messages []MessageItem
	var reply string

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

	if gptResponseBody != nil && len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Messages.Content
		log.Printf("gpt reply string %v\n", reply)
	}
	return reply, nil
}

func httpRequestImagesGenerations(msg string, runtimes int) (*ImageGPTResponseBody, error) {
	cfg := config.LoadConfig()
	if cfg.ApiKey == "" {
		return nil, errors.New("api key required")
	}

	requestBody := ImageGPTRequestBody{
		Prompt: msg,
		N:      cfg.ImageN,
		Size:   cfg.ImageSize,
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal requestBody error: %v", err)
	}

	log.Printf("gpt images request(%d) json: %s\n", runtimes, string(requestData))

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestData))
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

	gptResponseBody := &ImageGPTResponseBody{}
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal responseBody error: %v", err)
	}
	return gptResponseBody, nil
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
