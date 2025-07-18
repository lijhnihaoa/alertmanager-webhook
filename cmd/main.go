package main

import (
	alertmanager "alertmanagerWebhookAdapter/pkg/alertmanager"
	"flag"
)

/*
export FEISHU_WEBHOOK_1="https://open.feishu.cn/open-apis/bot/v2/hook/805d28c6-055d-4383-94bb-8e1d577d8587"
export FEISHU_WEBHOOK_2="https://open.feishu.cn/open-apis/bot/v2/hook/03b0a013-4b6b-447e-a1ee-7c68e9140c01"
*/

var (
	syslogProtocol = ""
)

func init() {

	flag.StringVar(&syslogProtocol, "syslog-protocol", "tcp", "syslog send protocol")
}

func main() {
	flag.Parse()

	alertmanager.Run(syslogProtocol)
}
