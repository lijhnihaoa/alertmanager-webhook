// Package syslogtools 提供了发送日志到 syslog 服务器的功能。
package syslogtools

import (
	"fmt"
	"log"
	"log/syslog"
	"strings"
	"unicode"
)

// Protocol 定义了发送 syslog 的协议，默认为 "tcp"。
var Protocol = "tcp"

// MaxSyslogMessageSize Syslog 单条消息最大长度（字节）
const MaxSyslogMessageSize = 1024

// sanitizeForSyslog 清理文本，移除 emoji 和非 ASCII 字符，确保 syslog 兼容。
func sanitizeForSyslog(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	for _, r := range text {
		// 只保留可打印的 ASCII 字符（32-126）和常见控制字符（换行、制表符）
		if r == '\n' || r == '\t' || (r >= 32 && r <= 126) {
			builder.WriteRune(r)
		} else if unicode.IsSpace(r) {
			builder.WriteRune(' ')
		}
		// 其他字符（emoji、中文等）直接忽略
	}

	return builder.String()
}

// splitMessage 将长消息分割成多个小块，每块不超过 maxSize 字节。
func splitMessage(text string, maxSize int) []string {
	// 先清理文本
	text = sanitizeForSyslog(text)

	// 如果消息不长，直接返回
	if len(text) <= maxSize {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		// 如果单行就超过限制，需要截断
		if len(line) > maxSize {
			line = line[:maxSize-20] + "...(truncated)"
		}

		// 检查添加这行后是否会超过限制
		if currentChunk.Len()+len(line)+1 > maxSize {
			// 当前块已满，保存并开始新块
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
	}

	// 添加最后一块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

func sendToSyslogServer(url, text string) error {
	// 连接到 syslog 服务，使用 LOG_LOCAL0 作为日志设施
	server, err := syslog.Dial(Protocol, url, syslog.LOG_ALERT|syslog.LOG_LOCAL0, "alertmanager-webhook")
	if err != nil {
		return fmt.Errorf("无法连接到 syslog: %w", err)
	}
	defer func() {
		if cerr := server.Close(); cerr != nil {
			log.Printf("failed to close syslog connection: %v", cerr)
		}
	}()

	// 分割消息
	chunks := splitMessage(text, MaxSyslogMessageSize)

	// 如果消息被分割了，添加序号
	if len(chunks) > 1 {
		for i, chunk := range chunks {
			msg := fmt.Sprintf("[Part %d/%d] %s", i+1, len(chunks), chunk)
			if err := server.Alert(msg); err != nil {
				return fmt.Errorf("发送日志失败 protocol: %s, url: %s, part: %d/%d: %w", Protocol, url, i+1, len(chunks), err)
			}
		}
		log.Printf("✅ Sent %d syslog messages (split from %d bytes)", len(chunks), len(text))
	} else {
		// 单条消息直接发送
		if err := server.Alert(chunks[0]); err != nil {
			return fmt.Errorf("发送日志失败 protocol: %s, url: %s: %w", Protocol, url, err)
		}
	}

	return nil
}
