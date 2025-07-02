FROM golang:1.24.3-alpine AS builder

              WORKDIR /feather

              COPY go.mod go.sum ./
              RUN go mod tidy && go mod download

              COPY . .

              RUN CGO_ENABLED=0 GOOS=linux go build -o main .

              FROM alpine:3.18

              WORKDIR /feather

              COPY --from=builder /feather/main .
              COPY config.toml .
              COPY assets/templates/argo ./assets/templates/argo

              RUN addgroup -S feather && adduser -S feather -G feather
              USER feather

              EXPOSE 8080

              CMD ["./main"]