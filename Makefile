gen:
	protoc --proto_path=proto --go_out=go/pb --go_opt=paths=source_relative --go-grpc_out=go/pb --go-grpc_opt=paths=source_relative proto/*.proto 

clean:
	rm go/pb/*

server: 
	cd go && go run cmd/server/main.go --port 8080

client: 
	cd go/cmd/client && go run . --address 0.0.0.0:8080

test:
	cd go && go test ./...