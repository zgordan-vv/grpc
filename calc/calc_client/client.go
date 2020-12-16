package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/zgordan-vv/grpc/calc/calcpb"
)

type server struct{}

func main() {
	fmt.Println("hello I am c client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Client: Failed to dial: %v", err)
	}
	defer conn.Close()

	c := calcpb.NewCalcServiceClient(conn)
	doSum(c)
	doDecomp(c)
	doAverage(c)
	doMax(c)
	doSqrt(c)
}

func doSum(c calcpb.CalcServiceClient) {
	fmt.Println("Unarysum")
	req := &calcpb.SumRequest{
		Sum: &calcpb.Sum{
			First: 1,
			Last: 10,
		},
	}
	resp, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Client: Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response:", resp.Result)
}

func doDecomp(c calcpb.CalcServiceClient) {
	fmt.Println("Decomp")
	req := &calcpb.DecompRequest{
		NumberToDecomp: 350901,
	}
	stream, err := c.Decomp(context.Background(), req)
	if err != nil {
		log.Fatalf("Client: Error while calling Greet RPC: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("Client: Error while reading stream: %v", err)
			}
		}
		log.Printf("Response:", resp.Result)
	}
}

func doAverage(c calcpb.CalcServiceClient) {
	fmt.Println("Average")
	reqs := []*calcpb.AverageRequest{
		&calcpb.AverageRequest{
			Next: 11,
		},
		&calcpb.AverageRequest{
			Next: 22,
		},
		&calcpb.AverageRequest{
			Next: 33,
		},
		&calcpb.AverageRequest{
			Next: 44,
		},
		&calcpb.AverageRequest{
			Next: 55,
		},
	}
	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Client: Error while calling Calc RPC: %v", err)
	}
	for _, req := range reqs {
		err := stream.Send(req)
		if err != nil {
			log.Fatalf("Client: Error while sending to Calc RPC: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Client: Error while receiving Calc RPC: %v", err)
	}
	fmt.Println("Received:", resp)
}

func doMax(c calcpb.CalcServiceClient) {
	fmt.Println("Max")
	reqs := []*calcpb.MaxRequest{
		&calcpb.MaxRequest{
			Next: 33,
		},
		&calcpb.MaxRequest{
			Next: 11,
		},
		&calcpb.MaxRequest{
			Next: 44,
		},
		&calcpb.MaxRequest{
			Next: 22,
		},
		&calcpb.MaxRequest{
			Next: 55,
		},
	}
	stream, err := c.Max(context.Background())
	if err != nil {
		panic(err)
	}
	waitc := make(chan struct{})
	go func(){
		for _, req := range reqs {
			err := stream.Send(req)
			if err != nil {
				panic(err)
			}
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
			fmt.Println("New max:", resp.Result)
		}
	}()
	<- waitc
}

func doSqrt(c calcpb.CalcServiceClient) {
	req := &calcpb.SquareRootRequest{
		Number: -10,
	}
	resp, err := c.SquareRoot(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Println("User error", respErr.Code(), respErr.Message())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("INVALID ARGUMENT")
			}
		} else {
			log.Fatalf("Client: Error while calling Sqrt RPC: %v", err)
		}
	} else {
		log.Printf("Response:", resp.Result)
	}
}
