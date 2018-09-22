FROM golang:1.11

RUN go get github.com/kudohamu/watchcat/...

CMD watchcat --conf=${CONFIG_PATH} --slack_webhook_url=${WEBHOOK_URL} --token=${GITHUB_TOKEN} --notifiers=std,slack w
