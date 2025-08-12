package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Message 定义了发送到飞书的消息结构。
type Message struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// NewMessage 创建一个新的飞书消息实例。
func NewMessage(text string) *Message {
	return &Message{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: text,
		},
	}
}

// SendToFeishu 发送消息到指定的飞书 webhook URL。
func (f *Message) SendToFeishu(webhookURL string, target string) error {
	body, _ := json.Marshal(f)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send to %s: %w", target, err)
	}

	// 检查 resp.Body.Close() 的错误
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	log.Printf("✅ Sent to %s: %s", target, resp.Status)
	return nil
}
