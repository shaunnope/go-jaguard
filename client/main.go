package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// flags
	addrLs     []string
	port       = flag.Int("port", 50000, "server port")
	joinAddr   = "localhost:50051,localhost:50052,localhost:50053,localhost:50054,localhost:50054,localhost:50055,localhost:50056"
	maxTimeout = flag.Int("maxTimeout", 100000, "max timeout for election")

	isRunningLocally = flag.Bool("l", false, "Set to true if running locally")
	host             = "localhost"
)

func listHelp() {
	fmt.Println("----------------------")
	fmt.Println("	OPTIONS:")
	fmt.Println("	ls path [-w] ")       //GetChildren
	fmt.Println("	get path [-w]")       // GetData
	fmt.Println("	getExists path [-w]") //GetExists

	fmt.Println("	create path [data]") //create node in path (acl not implemented)
	fmt.Println("	delete path")        //delete node in path (-v version flag not implemented)
	fmt.Println("	set path data")      //set data of node in path (-v version flag not implemented)
	fmt.Println("	sync")               //set data of node in path (-v version flag not implemented)

	fmt.Println("	q")
	fmt.Println("----------------------")
}
func parseReadCommand(command []string) (string, bool, error) {
	// Expect at least 2 elements in command (including the 'ls' itself)
	if len(command) < 2 {
		return "", false, fmt.Errorf("unknown argument '%s' for command", command)
	}

	path := command[1]
	setWatch := false

	// Check if there are more arguments, and if so, process them
	if len(command) > 2 {
		for _, arg := range command[2:] {
			switch arg {
			case "-w":
				setWatch = true
			default:
				return "", false, fmt.Errorf("unknown argument '%s' for command", arg)
			}
		}
	}
	return path, setWatch, nil
}

func menu() {
	input := bufio.NewScanner(os.Stdin)
	listHelp()
	fmt.Print("Enter your command: ")
Loop:
	for input.Scan() {
		command := strings.Split(input.Text(), " ")

		commandType := command[0]
		switch commandType {
		case "ls":
			fmt.Printf("Executing ls: %s\n", command)

			path, setWatch, err := parseReadCommand(command)
			if err != nil {
				log.Printf("%s\n", err)
				break
			}

			getChildrenReply, err := SendClientGrpc(pb.NodeClient.GetChildren, &pb.GetChildrenRequest{Path: path, SetWatch: setWatch, ClientHost: host, ClientPort: strconv.Itoa(*port)}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR LS: %s\n", err)
			} else {
				fmt.Printf("READ: %s has children: %s\n", path, getChildrenReply.Children)
			}

		case "get":
			fmt.Printf("Executing get: %s\n", &command)

			path, setWatch, err := parseReadCommand(command)
			if err != nil {
				log.Printf("%s\n", err)
				break
			}

			getData, err := SendClientGrpc(pb.NodeClient.GetData, &pb.GetDataRequest{Path: path, SetWatch: setWatch, ClientHost: host, ClientPort: strconv.Itoa(*port)}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR GET: %s\n", err)
				break
			}
			fmt.Printf("READ: %s %b\n", path, getData.Data)

		case "getExists":
			fmt.Printf("Executing getExists: %s\n", &command)

			path, setWatch, err := parseReadCommand(command)
			if err != nil {
				log.Printf("%s\n", err)
				break
			}

			getExists, err := SendClientGrpc(pb.NodeClient.GetExists, &pb.GetExistsRequest{Path: path, SetWatch: setWatch, ClientHost: host, ClientPort: strconv.Itoa(*port)}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR GET EXISTS: %s\n", err)
			} else {
				fmt.Printf("READ: %s exist: %t\n", path, getExists.Exists)
			}

		case "create":
			fmt.Printf("Executing create: %s\n", &command)

			if len(command) < 2 {
				fmt.Println("Not enough arguments for 'create' command")
				break
			}
			path := command[1]
			var data []byte
			setSequential := false
			setEphemeral := false

			switch len(command) {
			case 3:
				if command[2] == "-s" {
					setSequential = true
				} else if command[2] == "-e" {
					setEphemeral = true
				} else {
					data = []byte(command[2])
				}
			case 4:
				data = []byte(command[2])
				if command[3] == "-s" {
					setSequential = true
				}
			case 5:
				data = []byte(command[2])
				if command[3] == "-s" || command[4] == "-s" {
					setSequential = true
				}
				if command[3] == "-e" || command[4] == "-e" {
					setEphemeral = true
				}

			}
			CUDSResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: path, Data: []byte(data), Flags: &pb.Flag{IsSequential: setSequential, IsEphemeral: setEphemeral}, OperationType: pb.OperationType_WRITE}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR CREATE: %s\n", err)
			} else {
				fmt.Printf("WRITE: %s is accepted: %t, path: %s\n", path, *CUDSResponse.Accept, *CUDSResponse.Path)
			}

		case "set":
			fmt.Printf("Executing set: %s\n", &command)

			if len(command) < 3 {
				fmt.Println("not enough arguments for 'set' command")
				break
			}

			path := command[1]
			data := command[2]

			CUDSResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: path, Data: []byte(data), Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_UPDATE}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR SET: %s\n", err)
			} else {
				fmt.Printf("SET: %s is accepted: %t\n", path, *CUDSResponse.Accept)
			}

		case "delete":
			fmt.Printf("Executing delete: %s\n", &command)

			if len(command) < 2 {
				fmt.Println("not enough arguments for 'delete' command")
				break
			}

			path := command[1]

			CUDSResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: path, Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_DELETE}, *maxTimeout)

			if err != nil {
				fmt.Printf("ERROR DELETE: %s\n", err)
			} else {
				fmt.Printf("DELETE: %s is accepted: %t\n", path, *CUDSResponse.Accept)
			}

		case "sync":
			CUDSResponse, err := SendClientGrpc(pb.NodeClient.HandleClientCUDS, &pb.CUDSRequest{Path: "", Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_SYNC}, *maxTimeout)
			if err != nil {
				fmt.Printf("ERROR SYNC: %s\n", err)
			} else {
				fmt.Printf("SYNC: Accepted: %t\n", *CUDSResponse.Accept)
			}

		case "q":
			fmt.Println("Quiting...")
			break Loop
		default:
			fmt.Println("INVALID COMMAND ...")
			listHelp()
		}
		fmt.Print("Enter your command: ")
	}
}

func main() {
	// cli or file of commands to run
	// goroutines
	flag.Parse()

	var listeningIP string
	if !*isRunningLocally {
		host, _ = os.Hostname()
	}
	fmt.Printf("Host:%v\n", host)
	fmt.Printf("Port:%v\n", *port)

	listeningIP = fmt.Sprintf("%s:%d", host, *port)

	lis, err := net.Listen("tcp", listeningIP)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		fmt.Printf("Listening at: %v\n", listeningIP)
	}
	grpc_s := grpc.NewServer()
	client := Client{}
	pb.RegisterZkCallbackServer(grpc_s, &client)
	go grpc_s.Serve(lis)

	var addrLsStr string
	if !*isRunningLocally {
		addrLsStr = os.Getenv("ADDR")
	} else {
		addrLsStr = joinAddr
	}
	addrLs = strings.Split(addrLsStr, ",")
	rand.Shuffle(len(addrLs), func(i, j int) { addrLs[i], addrLs[j] = addrLs[j], addrLs[i] })
	fmt.Printf("Address list call order: %v\n", addrLs)

	menu()
}
func SendClientGrpc[T pb.Message, R pb.Message](
	F func(pb.NodeClient, context.Context, T, ...grpc.CallOption) (R, error),
	msg T,
	timeout int,
) (R, error) {
	var err error = nil
	var r R
	for _, serverAddr := range addrLs {
		conn, _ := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		defer conn.Close()

		c := pb.NewNodeClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err = F(c, ctx, msg)

		if err != nil {
			fmt.Printf("CONNECTION FAIL: %s\n", serverAddr)
			continue
		} else {
			fmt.Printf("CONNECTED: %s\n", serverAddr)
			break
		}
	}
	return r, err
}
