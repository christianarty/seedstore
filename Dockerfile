FROM golang:1.23
LABEL authors="christianarty"

WORKDIR /root
RUN mkdir "/root/bin"
VOLUME ["/data"]
ENV PS1="$(whoami)@$(hostname):$(pwd)\\$ " \
  HOME="/root" \
  TERM="xterm" \
  PATH="/root/bin:$PATH"


# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /root/bin/ ./...


CMD ["seedstore", "subscribe"]
