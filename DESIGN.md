# Design

## Structure
```.
.
├── build.sh
├── client
├── config.json
├── main.go
├── server
└── zouk
```
Jaguard has 3 main directories. Loosely speaking,  `client` and `server` deal with networking and read/write operations, while `zouk` deals with znodes, zookeeper protocol implementation and grpc interfaces.

## Issues