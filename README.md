

## 启动  go run cmd/main.go

- 配置 webhook 通过添加环境变量 FEISHU_WEBHOOK_xxx 来配置多个 webhook

export FEISHU_WEBHOOK_1="xxxx"
export FEISHU_WEBHOOK_2="xxxx"


- 告警配置可以通过 /feishu?target=ops,dev 配置 target=ops,dev 来实现控制警告发送给哪个 webhook

## Loki 日志查询功能（可选）

如果配置了 `LOKI_URL` 环境变量，adapter 会自动从 Loki 查询触发告警的实际日志内容，并包含在告警消息中。

### 配置方法

```bash
# 必需：Loki 服务地址
export LOKI_URL="http://loki:3100"

# 可选：Basic Auth 认证
export LOKI_USERNAME="xxx"
export LOKI_PASSWORD="xxx"

# 可选：查询参数（有默认值）
export LOKI_LOG_LIMIT="10"          # 返回最多 10 条日志
export LOKI_QUERY_RANGE="5"         # 查询最近 5 分钟的日志
export LOKI_QUERY_TIMEOUT="5s"      # 查询超时 5 秒
```

### Loki 规则配置

在 Loki 告警规则中添加 `log_query` 字段：

```yaml
annotations:
  summary: "【日志】Pod 出现 ERROR 日志"
  description: "容器出现错误日志"
  log_query: '{namespace="default", pod="my-pod"} |~ "(?i)(ERROR)"'  # 添加此字段
```

adapter 会使用 `log_query` 自动查询 Loki，获取实际触发告警的日志内容。

### 工作流程

1. Loki 规则触发告警 → Alertmanager → Webhook Adapter
2. Adapter 检测到 `log_query` 字段
3. 调用 Loki API 查询最近的匹配日志
4. 将日志内容格式化后附加到告警消息
5. 发送到飞书/Syslog

**注意**：
- 如果 Loki 查询失败，仍会发送告警（不会因为查询失败而阻止告警）
- 如果没有配置 `LOKI_URL`，功能自动禁用，不影响原有功能
- 日志内容有大小限制（飞书消息约 30KB），超过会被截断
