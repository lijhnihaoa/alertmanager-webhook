package syslogtools

import (
	"fmt"
	"log/syslog"
)

var Protocol = "tcp"

func sendToSyslogServer(url, text string) error {
	// 连接到本地 syslog 服务，使用 LOG_LOCAL0 作为日志设施
	server, err := syslog.Dial(Protocol, url, syslog.LOG_LOCAL0, "")
	if err != nil {
		return fmt.Errorf("无法连接到 syslog: %v", err)
	}
	defer server.Close() // 确保在程序结束时关闭连接

	// 发送信息到 syslog
	if err := server.Alert(text); err != nil {
		return fmt.Errorf("发送日志失败 protocol: %s, url: %s err: %v", Protocol, url, err)
	}
	return nil
}
