package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"github.com/zgordan-vv/grpc/greet/greetpb"
)

type server struct{}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Println("Server: greet function is invoked", req)
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	resp := greetpb.GreetResponse{
		Result: result,
	}
	return &resp, nil
}

func (*server) GreetDL(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Println("Server: greet function is invoked", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The client canceled the request")
			return nil, status.Error(codes.Canceled, "canceled")
		}
		time.Sleep(1 * time.Second)
	}
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	resp := greetpb.GreetResponse{
		Result: result,
	}
	return &resp, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Println("Server: greet stream function is invoked", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := fmt.Sprintf("Hello %s %d\n", firstName, i)
		resp := &greetpb.GreetManyTimesResponse {
			Result: result,
		}
		stream.Send(resp)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	result := ""
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
			fmt.Println("Server: will send", result)
				stream.SendAndClose(&greetpb.LongGreetResponse{
					Result: result,
				})
				break
			} else {
				log.Fatalf("Server: Error while reading from stream %v", err)
				return err
			}
		}
		fmt.Println("Server: received", req)
		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}
	return nil
}

func (*server)  GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("BiDi")
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				log.Fatalf("Server: Error while reading from stream %v", err)
				return err
			}
		}
		firstName := req.GetGreeting().GetFirstName()
		result := fmt.Sprintf("Hello %s !", firstName)
		err = stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	fmt.Println("hello world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	tls := false
	opts := []grpc.ServerOption{}

	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
		}
		opts = append(opts, grpc.Creds(creds))
	}
	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
