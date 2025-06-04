package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "example.com/grpc-chat-app/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Connection struct {
	pb.UnimplementedBroadcastServer
	stream pb.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Pool struct {
	pb.UnimplementedBroadcastServer
	Connection []*Connection
}

func (p *Pool) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	p.Connection = append(p.Connection, conn)
	log.Printf("User %s connected", conn.id)

	return <-conn.error
}

func (p *Pool) BroadcastMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	if msg.Timestamp == nil {
		msg.Timestamp = timestamppb.Now()
	}

	for _, conn := range p.Connection {
		wait.Add(1)

		go func(msg *pb.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {

				if err := conn.stream.Send(msg); err != nil {
					log.Printf("Error sending message to %s: %v", conn.id, err)

					conn.active = false
					conn.error <- err
				} else {
					log.Printf("Sent message to %s from %s", conn.id, msg.Id)
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &pb.Close{}, nil
}

func main() {
	grpcServer := grpc.NewServer()

	var connections []*Connection
	pool := &Pool{
		Connection: connections,
	}

	pb.RegisterBroadcastServer(grpcServer, pool)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error creating TCP listener: %v", err)
	}

	fmt.Println("gRPC Server started at port :8080")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error serving gRPC requests: %v", err)
	}
}
