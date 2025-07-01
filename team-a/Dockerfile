FROM golang:1.24.2-alpine AS builder
RUN apk add --no-cache upx
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-w -s' -o swarm
RUN upx --best --lzma swarm

FROM scratch
COPY --from=builder /app/swarm /swarm
ENTRYPOINT ["/swarm"]
CMD [ "run" ]