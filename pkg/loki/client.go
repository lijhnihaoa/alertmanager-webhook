// Package loki 提供与 Loki API 交互的客户端功能。
package loki

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client Loki API 客户端。
type Client struct {
	URL      string        // Loki 服务地址，如 http://loki:3100
	Username string        // Basic Auth 用户名（可选）
	Password string        // Basic Auth 密码（可选）
	Timeout  time.Duration // HTTP 请求超时时间
}

// QueryRangeResponse Loki query_range API 的响应结构。
type QueryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"` // [timestamp, log_line]
		} `json:"result"`
	} `json:"data"`
}

// QueryLogs 查询 Loki 日志，返回格式化的日志内容列表。
// query: LogQL 查询语句
// limit: 返回的最大日志条数
// rangeMinutes: 查询的时间范围（分钟）
func (c *Client) QueryLogs(query string, limit int, rangeMinutes int) ([]string, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("Loki URL not configured")
	}

	// 构建查询参数
	now := time.Now()
	end := now.UnixNano()
	start := now.Add(-time.Duration(rangeMinutes) * time.Minute).UnixNano()

	params := url.Values{}
	params.Add("query", query)
	params.Add("limit", strconv.Itoa(limit))
	params.Add("start", strconv.FormatInt(start, 10))
	params.Add("end", strconv.FormatInt(end, 10))
	params.Add("direction", "backward") // 从最新的日志开始

	// 构建完整 URL
	apiURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", c.URL, params.Encode())

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加 Basic Auth（如果配置了）
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// 设置超时
	client := &http.Client{
		Timeout: c.Timeout,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Loki: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Loki API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var queryResp QueryRangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取日志行
	var logs []string
	for _, result := range queryResp.Data.Result {
		for _, value := range result.Values {
			if len(value) >= 2 {
				// value[0] 是时间戳，value[1] 是日志内容
				logLine := value[1]
				logs = append(logs, logLine)

				// 限制返回的日志条数
				if len(logs) >= limit {
					return logs, nil
				}
			}
		}
	}

	return logs, nil
}

// FormatLogs 格式化日志列表为易读的文本。
// 会清理掉八进制转义序列（#011=tab, #012=newline）使输出更清晰。
func FormatLogs(logs []string, maxLines int) string {
	if len(logs) == 0 {
		return "（无日志内容）"
	}

	var builder strings.Builder
	count := min(len(logs), maxLines)

	for i := range count {
		// 清理日志中的八进制转义序列
		cleanLog := cleanLogLine(logs[i])
		fmt.Fprintf(&builder, "%d. %s | ", i+1, cleanLog)
	}

	if len(logs) > maxLines {
		fmt.Fprintf(&builder, "...（还有 %d 条日志未显示）", len(logs)-maxLines)
	}

	return builder.String()
}

// cleanLogLine 清理日志行中的转义序列和多余的空白字符。
func cleanLogLine(log string) string {
	// 替换常见的八进制转义序列
	log = strings.ReplaceAll(log, "#011", " ")  // Tab -> 空格
	log = strings.ReplaceAll(log, "#012", " ")  // 换行 -> 空格
	log = strings.ReplaceAll(log, "\t", " ")    // 实际 Tab -> 空格
	log = strings.ReplaceAll(log, "\n", " ")    // 实际换行 -> 空格
	log = strings.ReplaceAll(log, "\r", " ")    // 回车 -> 空格

	// 压缩多个连续空格为一个
	for strings.Contains(log, "  ") {
		log = strings.ReplaceAll(log, "  ", " ")
	}

	return strings.TrimSpace(log)
}
