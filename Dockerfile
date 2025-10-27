# Build stage
FROM golang:latest AS builder

WORKDIR /app
COPY checkhost.go .
#COPY go.mod go.sum* ./

# Initialize go module if go.mod doesn't exist
RUN if [ ! -f go.mod ]; then go mod init gatus-proxy; fi

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gatus-proxy .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/gatus-proxy .

EXPOSE 8080

CMD ["./gatus-proxy"]