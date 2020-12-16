package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"github.com/zgordan-vv/grpc/greet/greetpb"
)

type server struct{}

func main() {
	fmt.Println("hello I am c client")
	tls := false
	opts := grps.WithInsecure()
	if tls {
		certFile := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			panic(sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	}
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Client: Failed to dial: %v", err)
	}
	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)
	//fmt.Printf("created client: %f", c)
	doServerStreaming(c)
	doClientStreaming(c)
	doBiDiStreaming(c)
	doUnary(c)
	doUnaryDL(c, 5 * time.Second)
	doUnaryDL(c, 1 * time.Second)
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot0",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot1",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot2",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot3",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Client: Error creating stream %v", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending request %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error when receive response %v", err)
	}
	fmt.Println("LongRead response:", resp)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Unary")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName: "Maarek",
		},
	}
	resp, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Client: Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response:", resp.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Server streaming")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName: "Maarek",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Client: Error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		resp, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Client: Error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes:", resp.Result)
	}
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot0",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot1",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot2",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "robot3",
			},
		},
	}
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Client: Error while creating stream: %v", err)
	}
	waitc := make(chan struct{})
	go func(){
		for _, req := range requests {
			err := stream.Send(req)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Millisecond * 400)
		}
		stream.CloseSend()
	}()
	go func(){
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					close(waitc)
					break
				} else {
					panic(err)
				}
			}
			fmt.Println("RESP:", resp.Result)
		}
	}()
	<-waitc
}

func doUnaryDL(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Unary")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName: "Maarek",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := c.GreetDL(ctx, req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Println("User error", respErr.Code(), respErr.Message())
			if respErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Deadline was exceeded")
			} else {
				fmt.Println("unexpected")
			}
		} else {
			log.Fatalf("Client: Error while calling Sqrt RPC: %v", err)
		}
	} else {
		log.Printf("Response:", resp.Result)
	}
}
