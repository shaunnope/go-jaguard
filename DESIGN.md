# Design

- [Design](#design)
- [Structure](#structure)
- [Implementation](#implementation)
  - [Data Tree](#data-tree)
  - [ZooKeeper Atomic Broadcast (ZAB)](#zookeeper-atomic-broadcast-zab)
- [Issues](#issues)
- [References](#references)

# Structure
```
# top level directory
├── build.sh
├── client/
├── config.json
├── main.go
├── server/
└── zouk/
```
Jaguard comprises of 3 main directories. Loosely speaking, `client` and `server` deal with networking and read/write operations, while the `zouk` module defines all the types and interfaces for the ZooKeeper protocol such as znodes, zxids and the data tree. The grpc interfaces are defined in `zouk/zouk.proto`.

<details open><summary>Client</summary>

```
└── client_main.go
```
</details>

<details open><summary>Server</summary>

```
├── basic.go
├── client_cud.go
├── config.go
├── election.go
├── heartbeat.go
├── main.go
├── messages.go
├── server.go
├── state.go
├── vote.go
├── write.go
└── zab.go # zab protocol
```
</details>

<details open><summary>Zouk</summary>

```
├── datatree.go
├── event.go
├── messages.go
├── vote.go
├── watch.go
├── znode.go
├── zouk.pb.go # generated protobuf from proto
├── zouk.proto
├── zouk_grpc.pb.go # generated grpc protobuf from proto
└── zxid.go
```
</details>

# Implementation
## Data Tree
datatree is implemented as a hashmap, to optimise read operations to the leaves of the tree.

```
key       : value
tree path : pointer to node
```
## ZooKeeper Atomic Broadcast (ZAB)
ZAB is implemented as a 2-phase commit protocol. The leader sends a proposal to all followers, and waits for a quorum of ACKs before committing the proposal. The leader then sends a commit message to all followers, and waits for a quorum of ACKs before sending a commit message to the client. Our implementation is in `server/zab.go` and is adapted from the description of the protocol in the paper by [Medeiros](https://api.semanticscholar.org/CorpusID:14507005).

We also implemented the Fast Leader Election protocol described in the paper. The leader election protocol is implemented in `server/election.go`.


# Issues


# References
- Medeiros, A. (2012). ZooKeeper’s atomic broadcast protocol: Theory and practice.