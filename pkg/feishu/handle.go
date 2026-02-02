// Package feishu 提供处理飞书告警通知的功能。
package feishu

import (
	"alertmanagerWebhookAdapter/pkg/common"
	"alertmanagerWebhookAdapter/pkg/loki"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Handler 处理来自 Alertmanager 的 webhook 请求。
// 解析请求体中的 JSON 数据，并将告警信息发送到指定的飞书 webhook 地址。
// 如果请求中包含 target 参数，则只发送到指定的目标；
// 如果没有指定，则默认广播到所有已配置的飞书 webhook 地址。
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

	// 处理每个告警 - 单独发送到飞书，避免消息合并
	targetParam := r.URL.Query().Get("target")
	var targetWebhooks map[string]string

	if targetParam != "" {
		// 解析指定的目标
		targetWebhooks = make(map[string]string)
		targets := strings.Split(targetParam, ",")
		for _, t := range targets {
			t = strings.TrimSpace(strings.ToLower(t))
			if url, exists := common.FeishuWebhook[t]; exists {
				targetWebhooks[t] = url
			} else {
				log.Printf("⚠️ Target '%s' not found in configuration", t)
			}
		}
	} else {
		// 广播到所有配置的飞书
		targetWebhooks = common.FeishuWebhook
	}

	// 如果没有有效的目标，直接返回
	if len(targetWebhooks) == 0 {
		log.Println("⚠️ No valid feishu targets configured")
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
			alertName = "未知告警"
		}

		status := alert.Status
		if status == "" {
			status = "unknown"
		}

		summary := alert.Annotations["summary"]
		if summary == "" {
			summary = "无摘要信息"
		}

		desc := alert.Annotations["description"]
		if desc == "" {
			desc = "无详细描述"
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
						triggerLogs = fmt.Sprintf("（Loki 日志查询失败: %v）", err)
					}
				} else if len(logs) > 0 {
					// 查询成功，格式化日志内容
					formattedLogs := loki.FormatLogs(logs, common.LokiConfig.LogLimit)
					triggerLogs = formattedLogs
					log.Printf("✅ Queried %d logs from Loki for alert %s", len(logs), alertName)
				} else {
					// 查询成功但没有日志
					if triggerLogs == "" {
						triggerLogs = "（查询时间范围内无匹配日志）"
					}
				}
			}
		}

		builder.WriteString(fmt.Sprintf("🚨 *%s*\n状态: %s\n摘要: %s\n详情: %s\n",
			alertName, status, summary, desc))

		// 如果有触发日志信息，则添加显示
		if triggerLogs != "" {
			builder.WriteString(fmt.Sprintf("触发日志:\n%s\n", triggerLogs))
		}

		// 如果有 GeneratorURL，则显示
		if alert.GeneratorURL != "" {
			builder.WriteString(fmt.Sprintf("生成器: %s\n", alert.GeneratorURL))
		}

		text := builder.String()
		msg := NewMessage(text)

		// 发送到所有目标
		for name, webhookURL := range targetWebhooks {
			if err := msg.SendToFeishu(webhookURL, name); err != nil {
				log.Printf("❌ Failed to send alert %s to %s: %v", alertName, name, err)
			} else {
				log.Printf("✅ Sent alert %s to feishu %s", alertName, name)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Printf("❌ Failed to write response: %v", err)
	}
}
