package serializer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtobufToJSON converts protocol buffer message to JSON string
func ProtobufToJSON(message proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Indent:          "  ",
		UseProtoNames:   true,
		Multiline:       true,
	}

	msg, err := marshaler.Marshal(message)
	return string(msg), err
}

func JSONToMessage(data string, message proto.Message) error {
	return protojson.Unmarshal([]byte(data), message)
}
