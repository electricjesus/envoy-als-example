FROM golang:tip-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum main.go ./
RUN go mod download && go mod verify && go mod tidy && \
    go build -o main .


FROM scratch

COPY --from=builder /app/main /app/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 

EXPOSE 8080
CMD ["/app/main"]