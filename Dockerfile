# Build stage
FROM golang:latest AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o trakx ./cmd/trakx

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S trakx && adduser -S -G trakx -h /home/trakx trakx

WORKDIR /app
COPY --from=builder /build/trakx .

RUN mkdir -p /home/trakx/.config/trakx /home/trakx/.cache/trakx && \
    chown -R trakx:trakx /home/trakx /app

USER trakx

ENV HOME=/home/trakx

EXPOSE 1337/tcp
EXPOSE 1337/udp

CMD ["./trakx", "run"]
