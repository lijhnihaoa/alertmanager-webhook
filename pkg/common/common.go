package common

import (
	"log"
	"os"
	"strings"
	"time"
)

var FeishuWebhook = make(map[string]string)
var SyslogWebhook = make(map[string]string)

func LoadWebhooks() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "FEISHU_WEBHOOK_") {
			parts := strings.SplitN(env, "=", 2)
			key := strings.ToLower(strings.TrimPrefix(parts[0], "FEISHU_WEBHOOK_"))
			FeishuWebhook[key] = parts[1]
			continue
		}
		if strings.HasPrefix(env, "SYSLOG_WEBHOOK_") {
			parts := strings.SplitN(env, "=", 2)
			key := strings.ToLower(strings.TrimPrefix(parts[0], "SYSLOG_WEBHOOK_"))
			SyslogWebhook[key] = parts[1]
		}
	}
	log.Printf("🪝 Webhooks loaded:\n feishu webhook: %v\n syslog addresses: %v", FeishuWebhook, SyslogWebhook)
}

type WebhookMessage struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Status            string            `json:"status"` // "firing" or "resolved"
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`      // "firing" or "resolved"
	Labels       map[string]string `json:"labels"`      // 包含 alertname、severity、instance 等
	Annotations  map[string]string `json:"annotations"` // 包含 summary、description 等
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"` // Prometheus 生成 URL
	Fingerprint  string            `json:"fingerprint"`  // 唯一标识
}
