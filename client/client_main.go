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

			//TODO: Function to better parse Command
			path := command[1]
			setWatch := false
			if len(command) == 3 {
				optionInput := command[2]
				if optionInput == "-w" {
					setWatch = true
				}
			}

			getChildrenReply, err := SendClientGrpc[*pb.GetChildrenRequest, *pb.GetChildrenResponse](pb.NodeClient.GetChildren, &pb.GetChildrenRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			fmt.Printf("READ: %s has children: %s\n", path, getChildrenReply.Children)
			if err != nil {
				log.Printf("Error sending read request\n")
			}

		case "get":
			fmt.Printf("Executing get: %s\n", &command)

			path := command[1]
			setWatch := false
			if len(command) == 3 {
				optionInput := command[2]
				if optionInput == "-w" {
					setWatch = true
				}
			}

			getData, err := SendClientGrpc[*pb.GetDataRequest, *pb.GetDataResponse](pb.NodeClient.GetData, &pb.GetDataRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			fmt.Printf("READ: %s has data:%b\n", path, getData.Data)
			if err != nil {
				log.Printf("Error sending read request\n")
			}

		case "getExists":
			fmt.Printf("Executing getExists: %s\n", &command)

			path := command[1]
			setWatch := false
			if len(command) == 3 {
				optionInput := command[2]
				if optionInput == "-w" {
					setWatch = true
				}
			}

			getExists, err := SendClientGrpc[*pb.GetExistsRequest, *pb.GetExistsResponse](pb.NodeClient.GetExists, &pb.GetExistsRequest{Path: path, SetWatch: setWatch}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending read request: %s\n", err)
			} else {
				fmt.Printf("READ: %s exist: %t\n", path, getExists.Exists)
			}

		case "create":
			fmt.Printf("Executing create: %s\n", &command)

			path := command[1]
			data := command[2]
			// isSequential := false
			// if len(command) == 4 {
			// 	optionInput := command[3]
			// 	if optionInput == "-s" {
			// 		isSequential = true
			// 	}
			// }

			//TODO: Fix the flag data format
			createRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Data: []byte(data), Flags: "", OperationType: pb.OperationType_WRITE}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending create request: %s\n", err)
			} else {
				fmt.Printf("WRITE: %s is accepted: %t\n", path, *createRequest.Accept)
			}

		case "set":
			fmt.Printf("Executing set: %s\n", &command)

			path := command[1]
			data := command[2]
			// isSequential := false
			// if len(command) == 4 {
			// 	optionInput := command[3]
			// 	if optionInput == "-s" {
			// 		isSequential = true
			// 	}
			// }

			//TODO: Fix the flag data format
			setRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Data: []byte(data), Flags: "", OperationType: pb.OperationType_UPDATE}, *maxTimeout)

			if err != nil {
				log.Printf("Error sending set request: %s\n", err)
			} else {
				fmt.Printf("SET: %s is accepted: %t\n", path, *setRequest.Accept)
			}

		case "delete":
			fmt.Printf("Executing set: %s\n", &command)

			path := command[1]
			// isSequential := false
			// if len(command) == 4 {
			// 	optionInput := command[3]
			// 	if optionInput == "-s" {
			// 		isSequential = true
			// 	}
			// }

			//TODO: Fix the flag data format
			deleteRequest, err := SendClientGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, &pb.CUDRequest{Path: path, Flags: "", OperationType: pb.OperationType_DELETE}, *maxTimeout)

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
