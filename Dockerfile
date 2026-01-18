FROM golang:latest AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cas ./cmd/main

FROM alpine:latest
COPY --from=builder /build/cas /usr/local/bin/cas
RUN mkdir -p /data
WORKDIR /data
EXPOSE 8080
CMD ["sh", "-c", "cas init 2>/dev/null; cas serve --host 0.0.0.0 --port 8080"]