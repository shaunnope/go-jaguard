# Design

## Structure
```
# top level directory
├── build.sh
├── client
├── config.json
├── main.go
├── server
└── zouk
```
Jaguard has 3 main directories. Loosely speaking,  `client` and `server` deal with networking and read/write operations, while `zouk` deals with znodes, zookeeper protocol implementation and grpc interfaces.

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

## Issues