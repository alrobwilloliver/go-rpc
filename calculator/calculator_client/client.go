package main

import (
	"context"
	"fmt"
	calculatorpb "go-grpc-course/calculator/calculator_pb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doSum(c)

	// doPrimeNumber(c)

	// doComputeAverage(c)

	doClientBiDiStream(c)

}

func doSum(c calculatorpb.CalculatorServiceClient) {
	req := calculatorpb.SumRequest{
		Sum: &calculatorpb.Sum{
			FirstInt:  3,
			SecondInt: 10,
		},
	}

	res, err := c.Sum(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error while calling Sum RPC %v", err)
	}

	log.Printf("Response from Sum: %d", res.Response)
}

func doPrimeNumber(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 682356,
	}

	resStream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error calling PrimeNumberDecomposition RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error response from PrimeNumberDecomposition RPC: %v", err)
		}

		log.Printf("Response from PrimeNumberDecomposition: %d", msg.GetPrime().Prime)
	}
}

func doComputeAverage(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Client Streaming RPC...")

	requests := []*calculatorpb.ComputeAverageRequest{
		{
			Number: 10,
		},
		{
			Number: 3,
		},
		{
			Number: 5,
		},
		{
			Number: 6,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage %v", err)
	}

	for _, req := range requests {
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from ComputeAverage: %v", err)
	}

	fmt.Printf("ComputeAverage response: %v", res)
}

func doClientBiDiStream(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Client Streaming RPC...\n")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		numbers := []int64{4, 3, 6, 4, 12}
		for _, number := range numbers {
			fmt.Printf("Sending message %v\n", number)
			sendErr := stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			if sendErr != nil {
				log.Fatalf("Error when sending stream request: %v", sendErr)
			}
			time.Sleep(1000 * time.Millisecond)
		}
		closeErr := stream.CloseSend()
		if closeErr != nil {
			log.Fatalf("Error when closing stream: %v", closeErr)
		}
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}
			maximum := res.GetMaximum()
			log.Printf("Received: %v", maximum)
		}
		close(waitc)
	}()

	<-waitc
}
