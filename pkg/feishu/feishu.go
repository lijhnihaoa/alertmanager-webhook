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

// CardMessage 定义了飞书富文本卡片消息结构。
type CardMessage struct {
	MsgType string `json:"msg_type"`
	Card    struct {
		Header struct {
			Title struct {
				Tag     string `json:"tag"`
				Content string `json:"content"`
			} `json:"title"`
			Template string `json:"template"` // 颜色：red, orange, blue, grey等
		} `json:"header"`
		Elements []map[string]interface{} `json:"elements"`
	} `json:"card"`
}

// NewCardMessage 创建一个新的飞书卡片消息实例。
// title: 卡片标题
// color: 标题颜色 (red, orange, blue, grey等)
// infoText: 基本信息文本（Markdown格式）
// descriptionText: 详细描述文本
// logsText: 日志内容文本（Markdown格式）
func NewCardMessage(title, color, infoText, descriptionText, logsText string) *CardMessage {
	msg := &CardMessage{
		MsgType: "interactive",
	}

	// 设置标题
	msg.Card.Header.Title.Tag = "plain_text"
	msg.Card.Header.Title.Content = title
	msg.Card.Header.Template = color

	// 添加基本信息区块
	msg.Card.Elements = append(msg.Card.Elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]string{
			"tag":     "lark_md",
			"content": infoText,
		},
	})

	// 添加分隔线
	msg.Card.Elements = append(msg.Card.Elements, map[string]interface{}{
		"tag": "hr",
	})

	// 添加描述区块（如果有）
	if descriptionText != "" {
		msg.Card.Elements = append(msg.Card.Elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]string{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**详细描述**:\n%s", descriptionText),
			},
		})
	}

	// 添加日志内容区块
	if logsText != "" {
		msg.Card.Elements = append(msg.Card.Elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]string{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**触发日志（最近10条）**:\n%s", logsText),
			},
		})
	}

	return msg
}

// SendToFeishu 发送卡片消息到指定的飞书 webhook URL。
func (c *CardMessage) SendToFeishu(webhookURL string, target string) error {
	body, _ := json.Marshal(c)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send card to %s: %w", target, err)
	}

	// 检查 resp.Body.Close() 的错误
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	log.Printf("✅ Sent card to %s: %s", target, resp.Status)
	return nil
}
