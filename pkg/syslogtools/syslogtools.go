// Package syslogtools 提供了发送日志到 syslog 服务器的功能。
package syslogtools

import (
	"fmt"
	"log"
	"log/syslog"
)

// Protocol 定义了发送 syslog 的协议，默认为 "udp"。
// UDP 避免了 TCP octet counting 的 framing 问题
var Protocol = "udp"

// sendToSyslogServer 发送日志到指定的 syslog 服务器地址。
func sendToSyslogServer(url, text string) error {
	log.Printf("Protocol: %v, url:%v, text: %v", Protocol, url, text)
	// 连接到本地 syslog 服务，使用 LOG_LOCAL0 作为日志设施
	server, err := syslog.Dial(Protocol, url, syslog.LOG_LOCAL0, "")
	if err != nil {
		return fmt.Errorf("无法连接到 syslog: %w", err)
	}
	defer func() {
		if cerr := server.Close(); cerr != nil {
			log.Printf("failed to close syslog connection: %v", cerr)
		}
	}()

	// 发送信息到 syslog
	if err := server.Alert(text); err != nil {
		return fmt.Errorf("发送日志失败 protocol: %s, url: %s: %w", Protocol, url, err)
	}
	return nil
}
