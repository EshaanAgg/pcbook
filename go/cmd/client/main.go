package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eshaanagg/pcbook/go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
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

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("Cannot open image file: ", imagePath)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("Cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("Cannot send the image information in the stream: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("Can't send the chuck data to the server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Cannot close the stream and recieve response from the server: ", err)
	}

	log.Printf("Image uploaded to the server successfully with id: %s and size: %d", res.GetId(), res.GetSize())
}

func rateLaptop(laptopClient pb.LaptopServiceClient, laptopsIds []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot open stream to rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	// Go routine to recieve responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("No more responses")
				waitResponse <- nil
				return
			}

			if err != nil {
				waitResponse <- fmt.Errorf("cannot recieve stream response: %v", err)
				return
			}

			log.Print("Recieved response: ", res)
		}
	}()

	// Send requests
	for i, laptopId := range laptopsIds {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopId,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("Send request: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send in Rate Laptop: %v", err)
	}

	err = <-waitResponse
	return err
}

const (
	username        = "admin1"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
	const laptopServicePath = "/eshaanagg.pcbook.LaptopService/"

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
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

	authClient := NewAuthClient(conn, username, password)
	interceptor, err := NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("Cannot create auth interceptor: ", err)
	}

	connNew, err := grpc.Dial(
		*serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("Cannot dial server: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(connNew)
	testRateLaptop(laptopClient)
}
