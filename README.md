jaguard /ˈʤæˌɡɑːd/
- _noun_: A sentinel from the genus Panthera
- _verb_: past tense of jaga, to take care
- _noun_: A Go implementation of ZooKeeper

# Jaguard

A Go implementation of [Apache Zookeper](https://zookeeper.apache.org/) Protocol for 50.041 Distributed Systems based on the [Zookeeper paper](zookeeper.pdf) and open-source Java implementation of Apache Zookeeper.

- Zookeeper Client (CLI)
- Zookeeper Server Cluster

## Features
- Znode implementation 
- Znode read/write operations + replication across different servers for maintenance of data tree 
- Leader re-election protocol via Zookeeper Atomic Broadcast(ZAB)
- Ephemeral nodes
- File Watch
- Server recovery from previous snapshot 

## Guarantees
- Linearizable Write
- Wait-free Read: Fast reading from another non-leader node 
- Fault tolerance: Consistency despite adversarial conditions 
- FIFO client ordering

## Getting Started
### Pre-requsites
- Go 1.21
- grpc-go
### Build
The `*.pb.go` files are currently ignored by git. To generate them, run
```bash
$ ./build.sh
```
### Run
Then, run `make cli` to start the client and `make puppet` to start the Zookeper cluster.

## Design 
Read about the project structure, design considerations and issues in [DESIGN.md](DESIGN.md).

## Testing
- how to run

To run the servers for testing, you need to change the directory to `/server` and you can run with the following commands.

```shell
go run *.go -config=../config.json
```

It has a few flags for the Checkpoint 2 demonstration.

- `-multiple_req` : For multiple request through the same server
- `-multiple_cli` : For multiple request through different servers
- `-leader_verbo` : For more detail printing for leader election

The code progress as such

It start from `main.go`
- Parse the flag
- Initialise each server base on the config.json
    - By running `go Run(idx)`

`func Run(idx int)` is define in `server.go`. The function initalise the server and start the leader election by starting each node `go node.Serve(grpc_s)`

In `func (s *Server) Serve(grpc_s *grpc.Server)`, it will start the leader election and start a new Go routine for the heart beat. 

After the method `Serve`, the servers are ready for any request. To simulate the request we have another function called `func Simulate(s *Server, path string)` which can be found in `basic.go`.

To simulate the client request, the function will take in the desired server for the client. Then the function will craft the `ZabRequest` to send to the server.

```go
req := &pb.ZabRequest{
    Transaction: &pb.Transaction{
        Zxid:  s.LastZxid.Inc().Raw(),
        Path:  path,
        Data:  data,
        Type:  1,
        Flags: "someFlags",
    },
    RequestType: pb.RequestType_CLIENT,
}
_, err = c.SendZabRequest(ctx, req)
if err != nil {
    log.Printf("%d error sending zab request: %v", s.Id, err)
}
```

The logic to process all the different types of Zab messages can all be found in the `zab.go`. 

It has the details for the messages in the leader election and for the function `func (s *Server) SendZabRequest(ctx context.Context, in *pb.ZabRequest)`

It basically check for two criteria
- `isLeader := s.GetState() == LEADING` which is a boolean
- `switch in.RequestType` which can have the follow values
    - `pb.RequestType_PROPOSAL`
    - `pb.RequestType_ANNOUNCEMENT`
    - `pb.RequestType_CLIENT`

In `zab.go`, the messages are sent through the `SendGrpc` function. The function definition is in `messages.go`

Here is the function signature:
```go
func SendGrpc[T pb.Message, R pb.Message](
	F func(pb.NodeClient, context.Context, T, ...grpc.CallOption) (R, error),
	s *Server,
	to int,
	msg T,
	timeout int,
)
```

This should be sufficient to tell you the general flow of the programme and entry points.

## Acknowledgement
- Ivan Feng
- Joshua Ng
- Sean Yap
- Shi Hui
- Wai Shun