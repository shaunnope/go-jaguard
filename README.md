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

## Acknowledgement
- Ivan Feng
- Joshua Ng
- Sean Yap
- Shi Hui
- Wai Shun