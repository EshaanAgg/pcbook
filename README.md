# PCBook

A simple service that allows you to compare and get information about different laptops and systems.

This is made with the intend of learning about `Protobufs` and how they support both unidirectional and bidirectional streaming of messages and serialization. The microservice is made in `Go`. You can find some short notes on the theory of the implementation in the [notes](./notes/) directory.

### Makefile

We will be using a [`Makefile`](./Makefile) to manage all of the code generation and the starting of servers.

You can use the following commands:

- `make gen`: Generate all the relevant `proto` files for the Go microservice
- `make clean`: Delete all the auto-generated `proto` files for the Go microservice
- `make server`: Start the sever on port `8080`
- `make client`: Run the client script
- `make test`: Run all the tests in the project
