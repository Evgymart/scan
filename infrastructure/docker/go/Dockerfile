FROM golang:1.26.4-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache git && \
    addgroup -g 65532 nonroot && \
    adduser -D -u 65532 -G nonroot nonroot

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/api

FROM alpine
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/main /main

RUN mkdir /data && chown -R 65532:65532 /data
WORKDIR /data
USER 65532:65532
EXPOSE 80
ENTRYPOINT ["/main"]
