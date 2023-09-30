package service_test

import (
	"context"
	"testing"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/sample"
	"github.com/eshaanagg/pcbook/go/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoId := sample.NewLaptop()
	laptopNoId.Id = ""

	laptopInvalidId := sample.NewLaptop()
	laptopInvalidId.Id = "invalid_uuid"

	laptop := sample.NewLaptop()
	store := service.NewInMemoryLaptopStore()
	store.Save(laptop)

	testCases := []struct {
		name   string
		laptop *pb.Laptop
		store  service.LaptopStore
		code   codes.Code
	}{
		{
			name:   "Success with ID",
			laptop: sample.NewLaptop(),
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "Success with no ID",
			laptop: laptopNoId,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "Fail due to invalid ID",
			laptop: laptopInvalidId,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "Fail due to already exisiting ID",
			laptop: laptop,
			store:  store,
			code:   codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := service.NewLaptopServer(tc.store, nil, nil)

			res, err := server.CreateLaptop(context.Background(), req)

			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, st.Code())
			}
		})
	}
}
