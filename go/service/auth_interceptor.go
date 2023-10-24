package service

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a server interceptor for authentication and authorization
type AuthInterceptor struct {
	jwtManager      *JWTManager
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtManager *JWTManager, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{jwtManager, accessibleRoles}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> Unary interceptor: ", info.FullMethod)

		err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> Stream interceptor: ", info.FullMethod)

		err := interceptor.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := interceptor.accessibleRoles[method]
	if !ok {
		// Everyone can access these non listed methods
		return nil
	}

	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "Metadata is not provided in the request")
	}

	values := metadata["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "Authorization token is not provided in the request")
	}

	accessToken := values[0]
	user, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Access token is invalid: %v", err)
	}

	for _, role := range accessibleRoles {
		if role == user.Role {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}
