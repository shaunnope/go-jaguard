package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// flags
	addrLs []string
	port   = flag.Int("port", 50008, "server port")
	// addr       = flag.String("addr", "localhost:50056", "the address to connect to")
	joinAddr   = "localhost:50051,localhost:50052,localhost:50053,localhost:50054,localhost:50055,localhost:50056"
	maxTimeout = flag.Int("maxTimeout", 100000, "max timeout for election")

	isRunningLocally = flag.Bool("l", false, "Set to true if running locally")
	isLeader         = false
	electionRound    = 0
	nodeInQueue      = ""
	host             = "localhost"
)

func main() {
	flag.Parse()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		electionRound = 0
		isLeader = false
		attemptElection()
	}()

	// handle watch callbacks
	// setup zkcallback server

	var listeningIP string
	if !*isRunningLocally {
		host, _ = os.Hostname()
	}
	fmt.Printf("Host:%v\n", host)
	fmt.Printf("Port:%v\n", *port)

	listeningIP = fmt.Sprintf("%s:%d", host, *port)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		fmt.Printf("Listening at: %v\n", listeningIP)
	}
	grpc_s := grpc.NewServer()
	client := Client{}
	pb.RegisterZkCallbackServer(grpc_s, &client)
	go grpc_s.Serve(lis)

	sig := <-sigChan
	fmt.Println("Received signal:", sig)
	doGracefulShutdown()
	fmt.Println("Shutting down gracefully")
}

func doGracefulShutdown() {
	fmt.Println("Performing graceful shutdown tasks")
	fmt.Printf("Attempting to delete:{%s}", nodeInQueue)
	DeleteResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: nodeInQueue, Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_DELETE}, *maxTimeout)
	if err != nil {
		fmt.Printf("Error sending delete its own leader node: %s\n", err)
	} else {
		fmt.Printf("DELETE: %s is accepted: %t\n", nodeInQueue, *DeleteResponse.Accept)
	}
}

func SendClientGrpc[T pb.Message, R pb.Message](
	F func(pb.NodeClient, context.Context, T, ...grpc.CallOption) (R, error),
	msg T,
	timeout int,
) (R, error) {
	var err error = nil
	var r R

	var addrLsStr string

	// // Set up a connection to the server.
	// docker_addr := os.Getenv("ADDR")
	// if *isRunningLocally {
	// 	docker_addr = *addr
	// }

	if !*isRunningLocally {
		addrLsStr = os.Getenv("ADDR")
	} else {
		addrLsStr = joinAddr
	}
	addrLs = strings.Split(addrLsStr, ",")
	rand.Shuffle(len(addrLs), func(i, j int) { addrLs[i], addrLs[j] = addrLs[j], addrLs[i] })
	// fmt.Printf("Address list call order: %v\n", addrLs)
	// connectedServer := addrLs[rand.Intn(len(addrLs))]

	fmt.Printf("Client connect to Zookeeper Server at %s\n", addrLs[0])

	conn, err := grpc.Dial(addrLs[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNodeClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err = F(c, ctx, msg)
	return r, err
}

func attemptElection() {
	path := "/election_"
	data := fmt.Sprintf("localhost:%d", *port)
	setSequential := true
	setEphemeral := true

	CUDSResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: path, Data: []byte(data), Flags: &pb.Flag{IsSequential: setSequential, IsEphemeral: setEphemeral}, OperationType: pb.OperationType_WRITE}, *maxTimeout)

	if err != nil {
		log.Printf("Error sending create request: %s\n", err)
	} else {
		fmt.Printf("WRITE: %s is accepted: %t, path: %s\n", path, *CUDSResponse.Accept, *CUDSResponse.Path)
	}

	nodeInQueue = *CUDSResponse.Path

	// 2. Take the return path and find the sequence node before it and put a watch on it getExists if it exists, else go down the list
	if checkIfFirst(*CUDSResponse.Path) {
		isLeader = true
		fmt.Printf("I am the leader")
	}
}

func checkIfFirst(path string) bool {
	// Check if there is a election node that is before it, if yes, set a watch and break
	sequenceNumber, err := extractSequentialValue(path)
	if err != nil {
		fmt.Println("Error extracting X value:", err)
		log.Fatalf("Extracted invalid sequential number %s \n", path)
	}
	if !*isRunningLocally {
		host, _ = os.Hostname()
	}
	for i := sequenceNumber - 1; i >= 0; i-- {
		checkPath := fmt.Sprintf("/election_%010d", i)
		//edit host
		getExists, err := SendClientGrpc(pb.NodeClient.GetExists, &pb.GetExistsRequest{Path: checkPath, SetWatch: true, ClientHost: host, ClientPort: strconv.Itoa(*port)}, *maxTimeout)
		if err != nil {
			log.Printf("Error sending read request: %s\n", err)
		}
		if getExists.Exists {
			fmt.Printf("Znode with path: %s is ealier", checkPath)
			return false
		}
	}
	return true
}

func extractSequentialValue(path string) (int, error) {
	// Assuming the path is in the format "/node0_000000000X"
	parts := strings.Split(path, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid path format %s", path)
	}

	seqStr := strings.TrimPrefix(parts[1], "000000000")
	seqNumber, err := strconv.Atoi(seqStr)
	if err != nil {
		return 0, err
	}

	return seqNumber, nil
}
