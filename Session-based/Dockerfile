FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o session-based ./cmd

FROM alpine

WORKDIR /app

COPY --from=builder /app/nginx ./nginx
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/session-based .

CMD [ "./session-based" ]