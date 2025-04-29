FROM golang:1.23-bookworm AS builder

WORKDIR /golangProject

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH="/go/bin:${PATH}"

COPY . .

RUN mkdir -p docs/swagger
RUN swag init \
    --parseDependency \
    --parseInternal \
    --parseDepth 5 \
    -g internal/app/start/start.go \
    --output docs/swagger

WORKDIR /golangProject/cmd/app
RUN go build -o /main

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /main .

EXPOSE 8080

CMD ["./main"]