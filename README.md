# wechat-webhook-bot

build
```
docker build -t wechat-webhook-bot:v2 .
```

use
```
docker run --name wechat-webhook-bot \
  --restart always \
  -p 8000:8000 \
  -e WXWORK_WEBHOOK_BOT_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  -d wechat-webhook-bot:v2
```

use with k8s 
```
# modify wechat-webhook-bot.yaml
kubectl -n xxx apply -f wechat-webhook-bot.yaml
```

alertmanager example
```
  config:
    global:
      resolve_timeout: 5m
    route:
      group_by: ["job"]
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 12h
      receiver: 'wechat_alert'
      routes:
      - receiver: wechat_alert
        continue: true
    receivers:
    - name: wechat_alert
      webhook_configs:
      - send_resolved: true
        url: 'http://wechat-webhook-bot-srv:8000'
```

alert message example
```
[3]  未恢复的告警
Node kube-node-debian68 is Lost.
192.168.18.244 Disk is almost full (< 15% left) mountpoint /home
StatefulSet apisix-system/apisix-etcd has not matched the expected number of replicas for longer than 15 minutes.
[1]  已恢复的告警
Node kube-master-debian62 is Lost
```
