package feishu

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewFeishuMessage(text string) *FeishuMessage {
	return &FeishuMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: text,
		},
	}
}
func (f *FeishuMessage) SendToFeishu(webhookURL string, target string) error {

	body, _ := json.Marshal(f)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("✅ Sent to %s: %s", target, resp.Status)
	return nil
}
