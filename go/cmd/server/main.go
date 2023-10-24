package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/eshaanagg.pcbook.LaptopService/"

	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func seedUsers(userStore service.UserStore) error {
	err := service.CreateUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}

	return service.CreateUser(userStore, "user1", "secret", "user")
}

const (
	secretKey     = "SuperSecretKey123$"
	tokenDuration = 30 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("Start server on port: %v", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatalf("There was an error in adding the initial users to the store: %v", err)
	}
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopServer := service.NewLaptopServer(
		service.NewInMemoryLaptopStore(),
		service.NewDiskImageStore("serializer/tmp"),
		service.NewInMemoryRatingStore(),
	)

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}
	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalf("cannot start server: %v", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}
