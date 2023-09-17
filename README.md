# PCBook

A simple service that allows you to compare and get information about different laptops and systems.

This is made with the intend of learning about `Protobufs` and how they support both unidirectional and bidirectional streaming of messages and serialization.

The code has two micro-services constructed as a part of it which communicate with `protobuf`, one made in `Go` and the other in `Java`.

### Makefile

We will be using a [`Makefile`](./Makefile) to manage all of the code generation and the starting of servers.

You can use the following commands:

- `make gen`: Generate all the relevant `proto` files for the Go microservice.
- `make clean`: Delete all the auto-generated `proto` files for the Go microservice.
- `make run`: Start the go microservice.
