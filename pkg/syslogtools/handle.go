package syslogtools

import (
	"alertmanagerWebhookAdapter/pkg/common"
	"alertmanagerWebhookAdapter/pkg/loki"
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

	// 验证告警数量
	if len(payload.Alerts) == 0 {
		log.Println("⚠️ No alerts in payload")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Printf("❌ Failed to write response: %v", err)
		}
		return
	}

	// 处理每个告警 - 单独发送到 syslog，避免消息过大
	targetParam := r.URL.Query().Get("target")
	var targetAddrs map[string]string

	if targetParam != "" {
		// 解析指定的目标
		targetAddrs = make(map[string]string)
		targets := strings.Split(targetParam, ",")
		for _, t := range targets {
			t = strings.TrimSpace(strings.ToLower(t))
			if addr, exists := common.SyslogWebhook[t]; exists {
				targetAddrs[t] = addr
			} else {
				log.Printf("⚠️ Target '%s' not found in configuration", t)
			}
		}
	} else {
		// 广播到所有配置的 syslog
		targetAddrs = common.SyslogWebhook
	}

	// 如果没有有效的目标，直接返回
	if len(targetAddrs) == 0 {
		log.Println("⚠️ No valid syslog targets configured")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Printf("❌ Failed to write response: %v", err)
		}
		return
	}

	// 逐个处理告警
	for _, alert := range payload.Alerts {
		// 为每个告警构建消息
		var builder strings.Builder

		// 获取字段值，提供默认值
		alertName := alert.Labels["alertname"]
		if alertName == "" {
			alertName = "Unknown Alert"
		}

		status := alert.Status
		if status == "" {
			status = "unknown"
		}

		summary := alert.Annotations["summary"]
		if summary == "" {
			summary = "No summary"
		}

		desc := alert.Annotations["description"]
		if desc == "" {
			desc = "No description"
		}

		triggerLogs := alert.Annotations["trigger_logs"]

		// 尝试从 Loki 查询实际日志内容
		if common.LokiConfig.Enabled && common.LokiClient != nil {
			logQuery := alert.Annotations["log_query"]
			if logQuery != "" {
				logs, err := common.LokiClient.QueryLogs(
					logQuery,
					common.LokiConfig.LogLimit,
					common.LokiConfig.QueryRange,
				)
				if err != nil {
					log.Printf("⚠️ Failed to query Loki for alert %s: %v", alertName, err)
					// 查询失败时保留原有的 trigger_logs 或添加错误提示
					if triggerLogs == "" {
						triggerLogs = fmt.Sprintf("(Loki query failed: %v)", err)
					}
				} else if len(logs) > 0 {
					// 查询成功，格式化日志内容
					formattedLogs := loki.FormatLogs(logs, common.LokiConfig.LogLimit)
					triggerLogs = formattedLogs
					log.Printf("✅ Queried %d logs from Loki for alert %s", len(logs), alertName)
				} else {
					// 查询成功但没有日志
					if triggerLogs == "" {
						triggerLogs = "(No matching logs in query range)"
					}
				}
			}
		}

		builder.WriteString(fmt.Sprintf("Alert: %s | Status: %s | Summary: %s | Description: %s",
			alertName, status, summary, desc))

		// 如果有触发日志信息，则添加显示
		if triggerLogs != "" {
			builder.WriteString(fmt.Sprintf(" | Trigger Logs: %s", triggerLogs))
		}

		text := builder.String()

		// 发送到所有目标
		for name, syslogAddr := range targetAddrs {
			if err := sendToSyslogServer(syslogAddr, text); err != nil {
				log.Printf("❌ Failed to send alert %s to %s: %v", alertName, name, err)
			} else {
				log.Printf("✅ Sent alert %s to syslog %s", alertName, name)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Printf("❌ Failed to write response: %v", err)
	}
}
