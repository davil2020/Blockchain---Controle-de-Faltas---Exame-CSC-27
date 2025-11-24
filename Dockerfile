FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o node ./cmd/node

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/node /app/node

EXPOSE 8080

CMD ["/app/node"]
