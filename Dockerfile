FROM golang:1.23
LABEL authors="christianarty"
LABEL org.opencontainers.image.source="https://github.com/christianarty/seedstore"
LABEL org.opencontainers.image.description="Seedstore image for subscriptions"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /root
RUN mkdir "/root/bin"
VOLUME ["/data"]
ENV PS1="$(whoami)@$(hostname):$(pwd)\\$ " \
  HOME="/root" \
  TERM="xterm" \
  PATH="/root/bin:$PATH"

# Install mosquitto, mosquitto_sub, lftp, bash, nano, and openssh
RUN apt update && apt install -y mosquitto mosquitto-clients lftp bash nano openssh-server

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /root/bin/ ./...


CMD ["seedstore", "subscribe"]
