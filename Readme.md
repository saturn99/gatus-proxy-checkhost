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

