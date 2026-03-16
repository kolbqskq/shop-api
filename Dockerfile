FROM golang:1.25.6 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./cmd/main.go

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8000

CMD ["./main"]