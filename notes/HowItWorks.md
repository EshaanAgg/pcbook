# Introduction to gRPC

In gRPC, a client application can directly call a method on a server application on a different machine as if it were a local object, making it easier for you to create distributed applications and services. As in many RPC systems, gRPC is based around the idea of defining a service, specifying the methods that can be called remotely with their parameters and return types. On the server side, the server implements this interface and runs a gRPC server to handle client calls. On the client side, the client has a stub (referred to as just a client in some languages) that provides the same methods as the server.

## Protocol Buffers

Protocol buffer data is structured as `messages`, where each message is a small logical record of information containing a series of name-value pairs called `fields`. You can also define gRPC `services` in ordinary proto files, with RPC method parameters and return types specified as protocol buffer messages.

gRPC lets you define four kinds of service methods:

- A `simple RPC` where the client sends a request to the server using the stub and waits for a response to come back, just like a normal function call.
- A `server-side streaming RPC` where the client sends a request to the server and gets a stream to read a sequence of messages back. The client reads from the returned stream until there are no more messages.
- A `client-side streaming RPC` where the client writes a sequence of messages and sends them to the server, again using a provided stream. Once the client has finished writing the messages, it waits for the server to read them all and return its response. You specify a client-side streaming method by placing the stream keyword before the request type.
- A `bidirectional streaming RPC` where both sides send a sequence of messages using a read-write stream. The two streams operate independently, so clients and servers can read and write in whatever order they like: for example, the server could wait to receive all the client messages before writing its responses, or it could alternately read a message then write a message, or some other combination of reads and writes. The order of messages in each stream is preserved. You specify this type of method by placing the stream keyword before both the request and the response.
