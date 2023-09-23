package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// Set timeout on the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)

	// Handle the errors in the request
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Printf("Laptop already exists")
		} else {
			log.Fatalf("Cannot create laptop: %v", err)
		}
		return
	}

	log.Printf("A new laptop is created with the id: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("Search filter: %v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatalf("Cannot search laptop: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("cannot recieve messages: %v", err)
		}

		laptop := res.GetLaptop()
		log.Print("- Match found: ", laptop.GetId())
		log.Print("  + Brand: ", laptop.GetBrand())
		log.Print("  + Name: ", laptop.GetName())
		log.Print("  + CPU Cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("  + Min CPU GHZ: ", laptop.GetCpu().GetMinGhz())
		log.Print("  + RAM: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
		log.Print("  + Price: ", laptop.GetPriceUsd())
	}
}

func main() {
	serverAddress := flag.String("address", "", "The server port")
	flag.Parse()
	log.Printf("Start server on port: %v", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Cannot dial server: %v", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	// Create 10 random laptops
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3200,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	searchLaptop(laptopClient, filter)
}
