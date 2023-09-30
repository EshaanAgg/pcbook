package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20

// LaptopServer is the server that provides the laptop services
type LaptopServer struct {
	// Use a in-memory store instead of a database connection
	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
	// Embedded to have forward compatibility
	pb.UnimplementedLaptopServiceServer
}

func (server *LaptopServer) GetLaptopStore() LaptopStore {
	return server.laptopStore
}

// Returns a new LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
		ratingStore: ratingStore,
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
	if err := checkContextError(ctx); err != nil {
		return nil, err
	}

	err := server.laptopStore.Save(laptop)
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

	err := server.laptopStore.Search(
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

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	// Handling the first stream packet as a meta-data packet
	req, err := stream.Recv()
	if err != nil {
		log.Printf("Cannot recieve image info: %v", err)
		return status.Error(codes.Unknown, "cannot recieve image info")
	}

	laptopId := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()

	log.Printf("Recieved an upload image request for laptop %s with image type %s", laptopId, imageType)

	laptop, err := server.laptopStore.Find(laptopId)
	if err != nil {
		log.Printf("There was an error in searching for the laptop")
		return status.Errorf(codes.Internal, "cannot find laptop: %v", err)
	}
	if laptop == nil {
		log.Printf("No laptop with the provided id (%s) was found to be stored.", laptopId)
		return status.Errorf(codes.InvalidArgument, "No laptop exists with the id : %v", err)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	// Handle all the subsequent packets as image data packets
	for {
		if err := checkContextError(stream.Context()); err != nil {
			return err
		}

		log.Print("Waiting for chunk data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("Recieved all the image data")
			break
		}

		if err != nil {
			log.Printf("There was an error in recieving the stream data")
			return status.Errorf(codes.Unknown, "cannot recieve chunck data: %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		imageSize += size

		log.Printf("Recieved chunk with size: %d", size)

		if imageSize > maxImageSize {
			log.Print("The sent image is too large.")
			return status.Errorf(codes.InvalidArgument, "The send image is too large. The maximum upload limit is: %v", maxImageSize)
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			log.Print("Cannot append chunk to the image data")
			return status.Errorf(codes.Internal, "Cannot append the sent chunk to the image data: %v", err)
		}
	}

	imageId, err := server.imageStore.Save(laptopId, imageType, imageData)
	if err != nil {
		log.Print("There was an error in storing the image to the disk")
		return status.Errorf(codes.Internal, "Cannot save image to disk: %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageId,
		Size: uint32(imageSize),
	}

	// Return a response to the client and close the stream
	err = stream.SendAndClose(res)
	if err != nil {
		log.Fatal("Cannot close the stream.")
		return status.Errorf(codes.Internal, "Cannot close the stream and send response: %v", err)
	}

	log.Printf("The image is successfully saved with id: %s and size: %v", imageId, imageSize)
	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := checkContextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more data to recieve.")
			break
		}

		if err != nil {
			log.Fatalf("Cannot recieve stream request: %v", err)
			return status.Errorf(codes.Unknown, "Cannot recieve stream request: %v", err)
		}

		laptopId := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("Recieved a RateLaptop request with id: %v, score = %.2f", laptopId, score)

		found, err := server.laptopStore.Find(laptopId)
		if err != nil {
			log.Fatalf("LaptoreStore.Find() function failed: %v", err)
			return status.Errorf(codes.Internal, "LaptoreStore.Find() function failed: %v", err)
		}

		if found == nil {
			return status.Errorf(codes.NotFound, "There is no registered laptop with id: %v", laptopId)
		}

		rating := server.ratingStore.Add(laptopId, score)
		res := &pb.RateLaptopResponse{
			LaptopId:     laptopId,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}
		err = stream.Send(res)
		if err != nil {
			log.Fatalf("There was an error in sending the response: %v", err)
			return status.Errorf(codes.Internal, "There was an error in streaming the response to the client: %v", err)
		}
	}

	return nil
}

func checkContextError(ctx context.Context) error {
	if ctx.Err() == context.Canceled {
		log.Print("The request was cancelled")
		return status.Error(codes.Canceled, "The request was cancelled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		log.Print("Deadline for the request exceeded")
		return status.Error(codes.DeadlineExceeded, "Deadline exceeded")
	}

	return nil
}
