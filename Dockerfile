FROM golang:1.24.1-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Specifica il path del main package
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/fileserver

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/config/*.json ./config/

RUN ls -laR .

RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./api"]
