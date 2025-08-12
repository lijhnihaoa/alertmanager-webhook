// Package alertmanager 提供 Alertmanager webhook 适配器的核心功能。
package alertmanager

import (
	"alertmanagerWebhookAdapter/pkg/common"
	"alertmanagerWebhookAdapter/pkg/feishu"
	"alertmanagerWebhookAdapter/pkg/syslogtools"
	"log"
	"net/http"
	"time"
)

// Run 启动 Alertmanager webhook 适配器服务。
func Run(syslogProtocol string) {
	common.LoadWebhooks()
	if len(common.FeishuWebhook) == 0 && len(common.SyslogWebhook) == 0 {
		log.Fatal("❌ No FEISHU_WEBHOOK_xxx env vars found and no SYSLOG_WEBHOOK_xxx env vars found")
	}
	if len(common.FeishuWebhook) == 0 {
		log.Println("❌ No FEISHU_WEBHOOK_xxx env vars found")
	}
	http.HandleFunc("/feishu", feishu.Handler)

	if len(common.SyslogWebhook) == 0 {
		log.Println("❌ No SYSLOG_WEBHOOK_xxx env vars found")
	}
	syslogtools.Protocol = syslogProtocol
	http.HandleFunc("/syslog", syslogtools.Handler)

	log.Println("🚀 Multi-hook adapter is running on :8080")
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
