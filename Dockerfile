FROM golang:1.25-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bot ./cmd/bot

EXPOSE 8080

CMD ["./bot"]
