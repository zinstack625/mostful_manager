FROM golang:alpine

WORKDIR /usr/src/mostful-manager
# precache deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/mostful-manager &&\
    cp entrypoint.sh /usr/local/bin/entrypoint.sh &&\
    mkdir -p /etc/mostful-manager/ &&\
    cp config.json /etc/mostful-manager/config.json &&\
    rm -r /usr/src/mostful-manager

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
