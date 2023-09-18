package serializer_test

import (
	"os"
	"testing"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/sample"
	"github.com/eshaanagg/pcbook/go/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	currPath, err := os.Getwd()
	require.NoError(t, err)

	binaryPath := currPath + "/tmp/laptop.bin"
	jsonPath := currPath + "/tmp/laptop.json"

	laptop := sample.NewLaptop()

	err = serializer.WriteToBinaryFile(laptop, binaryPath)
	require.NoError(t, err)

	err = serializer.WriteToJSONFile(laptop, jsonPath)
	require.NoError(t, err)

	laptopRead := &pb.Laptop{}
	err = serializer.ReadFromBinaryFile(binaryPath, laptopRead)
	require.NoError(t, err)

	require.True(t, proto.Equal(laptop, laptopRead))
}
