jaguard /ˈʤæˌɡɑːd/
- _noun_: A sentinel from the genus Panthera
- _verb_: past tense of jaga, to take care
- _noun_: A Go implementation of ZooKeeper

> I jaguard my friends' belongings while they went zouk.

# Jaguard

A Go implementation of [Apache Zookeper](https://zookeeper.apache.org/) Protocol for 50.041 Distributed Systems based on the [Zookeeper paper](zookeeper.pdf) and open-source Java implementation of Apache Zookeeper.

Much like ZooKeeper, Jaguard is a wait-free coordination service that maintains a tree-like data structure. It is a fault-tolerant system that provides high-availability of reads, while guaranteeing linearizable writes.
## Table of Contents
- [Jaguard](#jaguard)
  - [Table of Contents](#table-of-contents)
- [Architecture](#architecture)
- [Features](#features)
  - [Znodes](#znodes)
  - [Watches](#watches)
- [Getting Started](#getting-started)
  - [Pre-requsites](#pre-requsites)
  - [Build](#build)
  - [Run](#run)
    - [Local](#local)
    - [Docker](#docker)
  - [Testing](#testing)
- [Example Use Case: Leader Election with Jaguard](#example-use-case-leader-election-with-jaguard)
- [Acknowledgements](#acknowledgements)

# Architecture
There are two main components to Jaguard: the **Jaguard clients** and the **Zouk servers**.

1. **Jaguard clients**: the user interface to the system. Clients can create, delete and set data on znodes, as well as get notifications when znodes are changed.
2. **Zouk servers**: the nodes that make up the distributed system. They maintain a sequentially consistent data tree and process client requests.


# Features
> Read about the project structure, design considerations and issues in [DESIGN.md](DESIGN.md).

Jaguard supports the following features:
- Linearizable Write: Writes are linearizable and atomic
- Wait-free Read: Fast reading from another non-leader node 
- Fault tolerance: Consistency despite adversarial conditions 
- FIFO client ordering: Requests from a client are executed in the order that they were sent

## Znodes
Znodes are the nodes that make up the data tree. They are similar to files and directories in a file system. Each znode has a path, data and a set of children znodes. Jaguard provides reliable read and write operations on znodes, and ensures a sequentially consistent replication of the data tree across all servers.

## Watches
Watches are notifications that are sent to clients when the data or state of a znode is changed. Clients can set watches on znodes to receive notifications when the znode is changed. Watches are one-time triggers, and are automatically removed after they are triggered.


# Getting Started
## Pre-requsites
- Go 1.21
- grpc-go
- protobuf

We assume that the working directory is the root of the project.

## Build
To generate the protobuf files, run
```shell
./build.sh
```
## Run
Jaguard can be run both locally and within Docker containers. While containers are recommended for demonstrating the fault-tolerance of the system, the functionality of the system can be emulated locally.

### Local
To run the servers locally, execute the following command.

```shell
go run ./server -local [-config=config.json] [-log=out] [-maxTimeout=1000]
```

- `-local` : Run the servers locally
- `-config` : Path to the config file. Defaults to `config.json`
- `-log` : Path to the directory for persistent storage. Defaults to `./out`
- `-maxTimeout` : Maximum timeout for messages. Defaults to `1000` ms

You can then interact with the servers using the client. To run a client, execute the following command.

```shell
go run ./client -l [-port=50000] [-maxTimeout=100000]
```

- `-l` : Run the client locally
- `-port` : Port to run the client on. Defaults to `50000`
- `-maxTimeout` : Maximum timeout for messages. Defaults to `100000` ms

### Docker
We have provided a docker-compose file for running the servers and clients in a Docker Compose cluster. To spin up the Docker Compose network, execute the following command.

```shell
docker compose up [--scale client=N] [--scale leader-elec-client=M]
```

To demonstrate the scalability of the system, we defined scalable services for the clients and leader election clients. By default, the Docker Compose network will spin up 1 client and 1 leader election client. However,
you can scale the number of clients and leader election clients by using the `--scale` flag.




## Testing
To run the servers for testing, execute the following command.

```shell
go run ./server -config=config.json
```
A few flags are defined for the demonstration in Checkpoint 2.

- `-multiple_req` : For multiple request through the same server
- `-multiple_cli` : For multiple request through different servers
- `-leader_verbo` : For more detail printing for leader election

The code progresses as such

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

# Example Use Case: Leader Election with Jaguard
In this example, we will demonstrate how Jaguard can be used to elect a leader within a group of clients. We will be using the Docker Compose network to demonstrate the fault-tolerance of the system.

1. Start the Docker Compose network with 3 or more leader election clients.
```shell
docker compose up --scale leader-elec-client=3
```
2. Start the leader election clients.
```shell
docker exec -it go-jaguard-leader-elec-client_1 go run client_main.go -l
```

- Leader election protocol via Zookeeper Atomic Broadcast (ZAB)

# Acknowledgements
- Ivan Feng
- Joshua Ng
- Sean Yap
- Shi Hui
- Wai Shun