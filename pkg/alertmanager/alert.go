package alertmanager

import (
	"alertmanagerWebhookAdapter/pkg/common"
	"alertmanagerWebhookAdapter/pkg/feishu"
	"alertmanagerWebhookAdapter/pkg/syslogtools"
	"fmt"
	"log"
	"net/http"
)

func Run(syslogProtocol string) {
	common.LoadWebhooks()
	if len(common.FeishuWebhook) == 0 && len(common.SyslogWebhook) == 0 {
		log.Fatal("❌ No FEISHU_WEBHOOK_xxx env vars found and no SYSLOG_WEBHOOK_xxx env vars found")
	}
	if len(common.FeishuWebhook) == 0 {
		fmt.Println("❌ No FEISHU_WEBHOOK_xxx env vars found")
	}
	http.HandleFunc("/feishu", feishu.Handler)

	if len(common.SyslogWebhook) == 0 {
		fmt.Println("❌ No SYSLOG_WEBHOOK_xxx env vars found")
	}
	syslogtools.Protocol = syslogProtocol
	http.HandleFunc("/syslog", syslogtools.Handler)

	log.Println("🚀 Multi-hook adapter is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
