package common

import (
	"log"
	"os"
	"strings"
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

type AlertmanagerPayload struct {
	Status string `json:"status"`
	Alerts []struct {
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	} `json:"alerts"`
}
