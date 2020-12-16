package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"github.com/zgordan-vv/grpc/calc/calcpb"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calcpb.SumRequest) (*calcpb.SumResponse, error) {
	fmt.Println("Server: sum function is invoked", req)
	first := req.GetSum().GetFirst()
	last := req.GetSum().GetLast()
	result := first + last
	resp := calcpb.SumResponse{
		Result: result,
	}
	return &resp, nil
}

func (*server) Decomp(req *calcpb.DecompRequest, stream calcpb.CalcService_DecompServer) error {
	numberToDecomp := int(req.NumberToDecomp)
	factor := 2
	for {
		if numberToDecomp == 1 {
			break
		}
		if numberToDecomp % factor == 0 {
			stream.Send(&calcpb.DecompResponse{
				Result: int32(factor),
			})
			numberToDecomp /= factor
			time.Sleep(200 * time.Millisecond)
		} else {
			factor++
		}
	}
	return nil
}

func (*server) Average(stream calcpb.CalcService_AverageServer) error {
	resultArr := []int32{}
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				avg := int32(0)
				if len(resultArr) != 0 {
					sum := int32(0)
					for _, n := range resultArr {
						sum += n
					}
					avg = sum / int32(len(resultArr))
				}
				stream.SendAndClose(&calcpb.AverageResponse{
					Result: avg,
				})
				break
			} else {
				log.Fatalf("Server: error when receiving a stream %v", err)
			}
		}
		resultArr = append(resultArr, req.Next)
	}
	return nil
}

func (*server) Max(stream calcpb.CalcService_MaxServer) error {
	max := int32(0)
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				log.Fatalf("Server: error when receiving a stream %v", err)
				return err
			}
		}
		if next := req.Next; next > max {
			max = next
			err := stream.Send(&calcpb.MaxResponse{
				Result: max,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (*server) SquareRoot(ctx context.Context, req *calcpb.SquareRootRequest) (*calcpb.SquareRootResponse, error) {
	number := req.Number
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received a negative number %f", number))
	}
	return &calcpb.SquareRootResponse{
		Result: math.Sqrt(number),
	}, nil
}

func main() {
	fmt.Println("hllo world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calcpb.RegisterCalcServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
