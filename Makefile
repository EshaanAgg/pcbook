gen:
	protoc --proto_path=proto --go_out=go/pb --go_opt=paths=source_relative --go-grpc_out=go/pb --go-grpc_opt=paths=source_relative proto/*.proto 

clean:
	rm go/pb/*

run: 
	cd go && go run main.go