FROM golang:1.26.4-alpine AS builder

WORKDIR /app
RUN addgroup -g 65532 nonroot && \
    adduser -D -u 65532 -G nonroot nonroot

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/api

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/main /main

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/main"]
