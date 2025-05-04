FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN apt-get update && apt-get install -y libpcap-dev
RUN CGO_ENABLED=1 GOOS=linux go build -o observer ./cmd/observer

FROM ubuntu:22.04
COPY --from=builder /app/observer /usr/bin/
COPY config/rules.yaml /etc/axom/
RUN apt-get update && apt-get install -y libpcap0.8 && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["observer"]