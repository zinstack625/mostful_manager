FROM golang:alpine

WORKDIR /usr/src/mostful-manager
# precache deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/mostful-manager

FROM alpine:latest
COPY --from=0 /usr/local/bin/mostful-manager /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/
COPY config.json.tmpl /etc/mostful-manager/config.json

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
