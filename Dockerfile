FROM golang:1.23 AS builder
WORKDIR /seedstore
COPY go.mod go.sum ./
RUN go mod download 
COPY . .
RUN CGO_ENABLED=0 go build -v -o ./cli/ ./...

FROM ghcr.io/linuxserver/baseimage-alpine:3.20 AS final

# Install additional packages
RUN apk add --no-cache mosquitto \
  mosquitto-clients \
  lftp \
  bash \
  nano \ 
  openssh \ 
  xz \
  coreutils \
  curl \
  findutils \
  jq \
  shadow 

RUN mkdir -p /root/.ssh && \
  echo -e "Host *\n\tStrictHostKeyChecking accept-new\n\n" > /root/.ssh/config && \
  chmod 600 /root/.ssh/config

COPY --from=builder /seedstore/cli /usr/local/bin
VOLUME /config

EXPOSE 1883 22

ENTRYPOINT [ "seedstore" ]
CMD [ "subscribe", "--configDir", "/config" ]