FROM golang:1.20 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY email_templates email_templates
CMD ["./main"]