FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/biocadtt ./src/cmd

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /out/biocadtt /usr/local/bin/biocadtt

RUN mkdir -p /app/reports /app/watch && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["biocadtt"]
