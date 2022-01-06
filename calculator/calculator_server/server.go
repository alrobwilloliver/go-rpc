package main

import (
	"context"
	"fmt"
	calculatorpb "go-grpc-course/calculator/calculator_pb"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("Called FindMaximum on the Server RPC as a streaming request \n")

	largest := int64(0)

	for {
		req, err := stream.Recv()

		if err != io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error receiving client FindMaximum RPC request: %v", err)
			return err
		}

		number := req.GetNumber()

		if number > largest {
			largest = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: number,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending the data to client: %v", sendErr)
				return sendErr
			}
		}
	}
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Called ComputeAverage on the Server RPC as a streaming request \n")
	total := int64(0)
	counter := 0
	result := int64(0)
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			result = total / int64(counter)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: float64(result),
			})
		}
		if err != nil {
			log.Fatalf("Error receiving client ComputeAverage RPC request: %v", err)
		}

		total += req.GetNumber()
		counter++
	}

}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	firstInt := req.GetSum().GetFirstInt()
	secondInt := req.GetSum().GetSecondInt()

	result := firstInt + secondInt
	res := &calculatorpb.SumResponse{
		Response: result,
	}
	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	number := req.GetNumber()
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			fmt.Printf("number %v \n", number)
			res := &calculatorpb.PrimeNumberDecompositionResponse{
				Prime: &calculatorpb.PrimeNumber{
					Prime: divisor,
				},
			}
			stream.Send(res)
			number = number / divisor
		} else {
			divisor++
			fmt.Printf("Divisor has increased to %v \n", divisor)
		}
	}

	return nil
}

func main() {
	fmt.Print("Running server \n")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
