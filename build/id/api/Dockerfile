FROM golang:1.21.1-alpine3.18 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o api_exec ./cmd/id/

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api_exec .
COPY /web/swagger ./web/swagger
COPY api/id ./api/id
COPY web/emails ./web/emails

EXPOSE 3002

CMD ["./api_exec"]