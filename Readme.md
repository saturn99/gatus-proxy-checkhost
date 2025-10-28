## gatus proxy 

It sends the link received from `gatus` the host to the `check-host.net` and sends the permanent link to the Telegram account.

## env
- TELEGRAM_TOKEN
- TELEGRAM_CHAT
- TELEGRAM_BASE (default: https://api.telegram.org/bot)
- PORT (default: 8080)
- PROXY_URL (optional)

## build:

### manual:
```
docker build -t gatus-proxy-checkhost:v2 .
```

### pre compile:
- go to https://github.com/saturn99/gatus-proxy-checkhost/releases
- download
- ``` chmod +x gatus-proxy-checkhost```
- run:
```
export TELEGRAM_TOKEN=xxx
export TELEGRAM_CHAT=xxx
./gatus-proxy-checkhost
```

### docker:
#### pull:
```
docker pull ghcr.io/saturn99/gatus-proxy-checkhost:v2
```
#### run:
```
docker run --rm \
    -p 8080:8080 \
    --env-file .env \
    ghcr.io/saturn99/gatus-proxy-checkhost:v2

```
OR
```bash
docker run --rm \
	-p 8080:8080 \
	-e TELEGRAM_TOKEN="" \
	-e TELEGRAM_CHAT= \
	-e TELEGRAM_BASE="https://api.telegram.org"  \
	ghcr.io/saturn99/gatus-proxy-checkhost:v2

```

#### compose:
```
IP=0.0.0.0 docker compose up
```

# Request:
```
curl -X POST \
  http://localhost:8080/webhook \
  -d '{
	"name":"https://www.yahoo.com/",
	"status":"check"
	}'
```



# use in gatus:

add alerting with custom

```
alerting:
  custom:
    url: "http://127.0.0.1:8080/webhook"
    method: "POST"
    body: |
      {
        "status": "[ALERT_TRIGGERED_OR_RESOLVED]",
        "name": "[ENDPOINT_URL]"
      }
    placeholders:
      ALERT_TRIGGERED_OR_RESOLVED:
        TRIGGERED: "check"
        RESOLVED: "OK"
```

- sure use `placeholders` with maped `TRIGGERED` to `check`


then use it to configs:

```
endpoints:
  - name: test
    url: https://test.test/
    alerts:
      - type: custom
        failure-threshold: 1
        send-on-resolved: true
```


### if host is down, app send check-host permanent link to your telegram