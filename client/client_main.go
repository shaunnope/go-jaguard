package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// flags
	port       = flag.Int("port", 50051, "server port")
	addr       = flag.String("addr", "localhost:50051", "the address to connect to")
	maxTimeout = flag.Int("maxTimeout", 100000, "max timeout for election")
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
	// Loop:
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

			getChildrenReply, err := SendClientGrpc[*pb.GetChildrenRequest, *pb.GetChildrenResponse](pb.NodeClient.GetChildren, &pb.GetChildrenRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			fmt.Printf("READ: %s has children: %s\n", path, getChildrenReply.Children)
			if err != nil {
				log.Printf("Error sending read request\n")
			}

		case "get":
			fmt.Printf("Executing get: %s\n", &command)

			path, setWatch, err := parseReadCommand(command)
			if err != nil {
				log.Printf("%s\n", err)
				break
			}

			getData, err := SendClientGrpc[*pb.GetDataRequest, *pb.GetDataResponse](pb.NodeClient.GetData, &pb.GetDataRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			fmt.Printf("READ: %s has data:%b\n", path, getData.Data)
			if err != nil {
				log.Printf("Error sending read request\n")
			}

		case "getExists":
			fmt.Printf("Executing getExists: %s\n", &command)

			path, setWatch, err := parseReadCommand(command)
			if err != nil {
				log.Printf("%s\n", err)
				break
			}

			getExists, err := SendClientGrpc[*pb.GetExistsRequest, *pb.GetExistsResponse](pb.NodeClient.GetExists, &pb.GetExistsRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending read request: %s\n", err)
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
			createRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Data: []byte(data), Flags: &pb.Flag{IsSequential: setSequential, IsEphemeral: setEphemeral}, OperationType: pb.OperationType_WRITE}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending create request: %s\n", err)
			} else {
				fmt.Printf("WRITE: %s is accepted: %t\n", path, *createRequest.Accept)
			}

		case "set":
			fmt.Printf("Executing set: %s\n", &command)

			if len(command) < 3 {
				fmt.Println("not enough arguments for 'set' command")
				break
			}

			path := command[1]
			data := command[2]

			setRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Data: []byte(data), Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_UPDATE}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending set request: %s\n", err)
			} else {
				fmt.Printf("SET: %s is accepted: %t\n", path, *setRequest.Accept)
			}

		case "delete":
			fmt.Printf("Executing set: %s\n", &command)

			if len(command) < 2 {
				fmt.Println("not enough arguments for 'delete' command")
				break
			}

			path := command[1]

			deleteRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Flags: &pb.Flag{IsSequential: false, IsEphemeral: false}, OperationType: pb.OperationType_DELETE}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending delete request: %s\n", err)
			} else {
				fmt.Printf("DELETE: %s is accepted: %t\n", path, *deleteRequest.Accept)
			}

		case "q":
			fmt.Println("Quiting...")
			break Loop
		default:
			fmt.Println("INVALID COMMAND ...")
			listHelp()
		}
	}
}

func main() {
	// cli or file of commands to run
	// goroutines
	flag.Parse()
	fmt.Printf("%v\n", port)
	menu()
}
func SendClientGrpc[T pb.Message, R pb.Message](
	F func(pb.NodeClient, context.Context, T, ...grpc.CallOption) (R, error),
	msg T,
	timeout int,
) (R, error) {
	var err error = nil
	var r R

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
