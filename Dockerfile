FROM golang:1.24-alpine AS build

WORKDIR /app

RUN apt-get update && apt-get install -y screen

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/bot .

CMD ["./bot"]