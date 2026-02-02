// Package common 提供了 Alertmanager Webhook 适配器的通用功能和数据结构。
package common

import (
	"alertmanagerWebhookAdapter/pkg/loki"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// FeishuWebhook 存储所有可用的飞书 webhook 地址，key 为目标标识。
var FeishuWebhook = make(map[string]string)

// SyslogWebhook 存储所有可用的 syslog webhook 地址，key 为目标标识。
var SyslogWebhook = make(map[string]string)

// LokiClient Loki API 客户端实例（全局单例）。
var LokiClient *loki.Client

// LokiConfig Loki 配置参数。
var LokiConfig struct {
	Enabled      bool          // 是否启用 Loki 查询功能
	LogLimit     int           // 返回的最大日志条数
	QueryRange   int           // 查询时间范围（分钟）
	QueryTimeout time.Duration // 查询超时时间
}

// LoadWebhooks 从环境变量中加载所有的 Webhook 配置和 Loki 配置。
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

	// 加载 Loki 配置
	loadLokiConfig()

	log.Printf("🪝 Webhooks loaded:\n feishu webhook: %v\n syslog addresses: %v\n loki enabled: %v",
		FeishuWebhook, SyslogWebhook, LokiConfig.Enabled)
}

// loadLokiConfig 从环境变量加载 Loki 配置。
func loadLokiConfig() {
	lokiURL := os.Getenv("LOKI_URL")
	if lokiURL == "" {
		log.Println("⚠️ LOKI_URL not set, Loki log query disabled")
		LokiConfig.Enabled = false
		return
	}

	// 设置默认值
	LokiConfig.Enabled = true
	LokiConfig.LogLimit = 10
	LokiConfig.QueryRange = 5
	LokiConfig.QueryTimeout = 5 * time.Second

	// 从环境变量读取自定义配置
	if limit := os.Getenv("LOKI_LOG_LIMIT"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			LokiConfig.LogLimit = val
		}
	}

	if rangeMinutes := os.Getenv("LOKI_QUERY_RANGE"); rangeMinutes != "" {
		if val, err := strconv.Atoi(rangeMinutes); err == nil && val > 0 {
			LokiConfig.QueryRange = val
		}
	}

	if timeout := os.Getenv("LOKI_QUERY_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			LokiConfig.QueryTimeout = val
		}
	}

	// 初始化 Loki 客户端
	LokiClient = &loki.Client{
		URL:      lokiURL,
		Username: os.Getenv("LOKI_USERNAME"),
		Password: os.Getenv("LOKI_PASSWORD"),
		Timeout:  LokiConfig.QueryTimeout,
	}

	log.Printf("✅ Loki client initialized: URL=%s, Limit=%d, Range=%dm, Timeout=%v",
		lokiURL, LokiConfig.LogLimit, LokiConfig.QueryRange, LokiConfig.QueryTimeout)
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
