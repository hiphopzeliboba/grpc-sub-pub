FROM golang:1.23 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o grpc_server internal/cmd/grpc_server/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/grpc_server /app/
COPY --from=builder /app/.env /app/.env
EXPOSE 50051

CMD ["./grpc_server"]