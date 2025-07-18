package feishu

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"alertmanagerWebhookAdapter/pkg/common"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var payload common.AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var text string
	for _, alert := range payload.Alerts {
		alertName := alert.Labels["alertname"]
		status := alert.Status
		summary := alert.Annotations["summary"]
		desc := alert.Annotations["description"]
		text += fmt.Sprintf("🚨 *%s*\n状态: %s\n摘要: %s\n详情: %s\n\n", alertName, status, summary, desc)
	}

	msg := NewFeishuMessage(text)
	targetParam := r.URL.Query().Get("target")
	if targetParam != "" {
		targets := strings.Split(targetParam, ",")
		for _, t := range targets {
			t = strings.TrimSpace(strings.ToLower(t))
			if err := msg.SendToFeishu(common.FeishuWebhook[t], t); err != nil {
				log.Printf("❌ Failed to send to %s: %v", t, err)
			}
		}
	} else {
		// 默认广播全部
		for _, v := range common.FeishuWebhook {
			msg.SendToFeishu(v, v)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
