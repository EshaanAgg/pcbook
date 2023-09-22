package service_test

import (
	"context"
	"net"
	"testing"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/sample"
	"github.com/eshaanagg/pcbook/go/serializer"
	"github.com/eshaanagg/pcbook/go/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopSever, serverAddress := startTestLatopServer(t)
	laptopClient := startTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedId := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedId, res.Id)

	// Check that the laptop is stored on the server (store)
	fetchedLaptop, err := laptopSever.Store.Find(laptop.Id)
	require.NoError(t, err)
	require.NotNil(t, fetchedLaptop)
	ensureSameLaptop(t, fetchedLaptop, laptop)
}

func startTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func startTestLatopServer(t *testing.T) (*service.LaptopServer, string) {
	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0") // Assign it any random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener) // Blocking code, start in a separate goroutine

	return laptopServer, listener.Addr().String()
}

// Serializes both the laptops to JSON and then compares them so that we can neglect the internally added fields by GRPC
func ensureSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
