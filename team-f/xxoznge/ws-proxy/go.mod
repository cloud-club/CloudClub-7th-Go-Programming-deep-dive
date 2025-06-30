module ws-proxy

go 1.24.4

require (
	example.com/golang-grpc-chat v0.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.5.3
	google.golang.org/grpc v1.73.0
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace example.com/golang-grpc-chat => ../golang-grpc-chat
