

## 启动  go run cmd/main.go

- 配置 webhook 通过添加环境变量 FEISHU_WEBHOOK_xxx 来配置多个 webhook

export FEISHU_WEBHOOK_1="xxxx"
export FEISHU_WEBHOOK_2="xxxx"


- 告警配置可以通过 /feishu?target=ops,dev 配置 target=ops,dev 来实现控制警告发送给哪个 webhook
