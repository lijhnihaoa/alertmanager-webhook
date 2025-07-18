

## 启动  go run cmd/main.go

- 配置 webhook 通过添加环境变量 FEISHU_WEBHOOK_xxx 来配置多个 webhook

export FEISHU_WEBHOOK_1="https://open.feishu.cn/open-apis/bot/v2/hook/805d28c6-055d-4383-94bb-8e1d577d8587"
export FEISHU_WEBHOOK_2="https://open.feishu.cn/open-apis/bot/v2/hook/03b0a013-4b6b-447e-a1ee-7c68e9140c01"


- 告警配置可以通过 /feishu?target=ops,dev 配置 target=ops,dev 来实现控制警告发送给哪个 webhook
