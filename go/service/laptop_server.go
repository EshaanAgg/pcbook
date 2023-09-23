package service

import (
	"context"
	"errors"
	"log"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer is the server that provides the laptop services
type LaptopServer struct {
	// Use a in-memory store instead of a database connection
	Store LaptopStore
	// Embedded to have forward compatibility
	pb.UnimplementedLaptopServiceServer
}

// Returns a new LaptopServer
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
	}
}

// The server must implement the LaptopServiceServer interface
// It is a unary RPC to create a new laptop
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("Recieved a CreateLaptop request with id: %s", laptop.Id)

	// Check if the client has sent an ID
	if len(laptop.Id) > 0 {
		// Check if the sent UUID is valid
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Cannot generate a new laptop id: %v", err)
		}
		laptop.Id = id.String()
	}

	// Check for context errors before saving the laptop to the store
	if ctx.Err() == context.Canceled {
		log.Print("The request was cancelled")
		return nil, status.Error(codes.Canceled, "The request was cancelled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		log.Print("Deadline for the request exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "Deadline exceeded")
	}

	err := server.Store.Save(laptop)
	if err != nil {
		// Figure out the appropiate error code
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "Cannot save laptop to the store: %v", err)
	}

	log.Printf("Saved laptop with the id: %v", laptop.Id)

	return &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}, nil
}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("Recieved a SearchLaptop request with filter: %v", filter)

	err := server.Store.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{
				Laptop: laptop,
			}

			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("Sent laptop with id: %v", laptop.Id)
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}
