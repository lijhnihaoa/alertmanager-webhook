package syslogtools

import (
	"alertmanagerWebhookAdapter/pkg/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Handler 处理来自 Alertmanager 的 syslog 请求。
// 解析请求体中的 JSON 数据，并将告警信息发送到指定的 syslog 地址。
// 如果请求中包含 target 参数，则只发送到指定的目标；
// 如果没有指定，则默认广播到所有已配置的 syslog 地址。
func Handler(w http.ResponseWriter, r *http.Request) {
	var payload common.WebhookMessage
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
		text += fmt.Sprintf("🚨 *%s*\n状态: %s\n摘要: %s\n详情: %s\n故障集群:%s\n\n",
			alertName, status, summary, desc, alert.GeneratorURL)
	}

	targetParam := r.URL.Query().Get("target")
	if targetParam != "" {
		targets := strings.Split(targetParam, ",")
		for _, t := range targets {
			t = strings.TrimSpace(strings.ToLower(t))
			if err := sendToSyslogServer(common.SyslogWebhook[t], text); err != nil {
				log.Printf("❌ Failed to send to %s: %v", t, err)
			}
		}
	} else {
		// 默认广播全部
		for _, v := range common.SyslogWebhook {
			err := sendToSyslogServer(v, text)
			if err != nil {
				log.Printf("❌ Failed to send to %s: %v", v, err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Printf("❌ Failed to write response: %v", err)
	}
}
