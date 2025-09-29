FROM golang:1.25-alpine AS builder

WORKDIR /aiproxy-free

COPY ./ /aiproxy-free

RUN go build -trimpath -ldflags "-s -w" -o aiproxy-free

FROM alpine:latest

RUN mkdir -p /aiproxy-free

WORKDIR /aiproxy-free

VOLUME /aiproxy-free

RUN apk add --no-cache ca-certificates tzdata ffmpeg curl && \
    rm -rf /var/cache/apk/*

COPY --from=builder /aiproxy-free/aiproxy-free /usr/local/bin/aiproxy-free

ENV PUID=0 PGID=0 UMASK=022

ENV FFMPEG_ENABLED=true

EXPOSE 3000

ENTRYPOINT ["aiproxy-free"]
