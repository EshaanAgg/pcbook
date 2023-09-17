gen:
	protoc --proto_path=proto --go_out=go/protobuf --go_opt=paths=source_relative proto/*.proto 