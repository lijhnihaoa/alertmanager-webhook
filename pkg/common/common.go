// Package common 提供了 Alertmanager Webhook 适配器的通用功能和数据结构。
package common

import (
	"log"
	"os"
	"strings"
	"time"
)

// FeishuWebhook 存储所有可用的飞书 webhook 地址，key 为目标标识。
var FeishuWebhook = make(map[string]string)

// SyslogWebhook 存储所有可用的 syslog webhook 地址，key 为目标标识。
var SyslogWebhook = make(map[string]string)

// LoadWebhooks 从环境变量中加载所有的 Webhook 配置。
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

// WebhookMessage 定义了 Alertmanager 发送的 webhook 消息格式。
// 该结构体包含了所有必要的字段，用于解析和处理 Alertmanager 的 webhook 消息。
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

// Alert 定义了单个告警的结构体，包含了告警的状态、标签、注释等信息。
// 该结构体用于解析 Alertmanager 发送的告警信息。
// 包含了告警的状态、标签、注释、开始时间、结束时间、生成 URL 和唯一标识等字段。
type Alert struct {
	Status       string            `json:"status"`      // "firing" or "resolved"
	Labels       map[string]string `json:"labels"`      // 包含 alertname、severity、instance 等
	Annotations  map[string]string `json:"annotations"` // 包含 summary、description 等
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"` // Prometheus 生成 URL
	Fingerprint  string            `json:"fingerprint"`  // 唯一标识
}
