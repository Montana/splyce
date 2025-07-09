FROM golang:1.22 as builder

WORKDIR /app
COPY main.go .
RUN go build -o splyce main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/splyce .

EXPOSE 8125/udp
EXPOSE 9100
ENTRYPOINT ["./splyce"]
