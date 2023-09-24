package service_test

import (
	"context"
	"io"
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

	laptopSever, serverAddress := startTestLatopServer(t, service.NewInMemoryLaptopStore())
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
	fetchedLaptop, err := laptopSever.GetLaptopStore().Find(laptop.Id)
	require.NoError(t, err)
	require.NotNil(t, fetchedLaptop)
	ensureSameLaptop(t, fetchedLaptop, laptop)
}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_MEGABYTE},
	}

	store := service.NewInMemoryLaptopStore()
	expectedIds := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIds[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIds[laptop.Id] = true
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

	_, serverAddress := startTestLatopServer(t, store)
	laptopClient := startTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIds, res.GetLaptop().GetId())
		found += 1
	}

	require.Equal(t, found, len(expectedIds))
}

func startTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func startTestLatopServer(t *testing.T, laptopStore *service.InMemoryLaptopStore) (*service.LaptopServer, string) {
	laptopServer := service.NewLaptopServer(laptopStore, nil)

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
