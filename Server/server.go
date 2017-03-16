package main

import (
	"io"
	"log"
	"net"
	"os/exec"

	pb "github.com/taylorflatt/go-chat"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":12021"
)

// Server is used to implement the RemoteCommandServer
type server struct{}

// List of clients connected to the server. Here ip is key and port is value.
var clients map[string]int32

// Executes a remote command from a client and returns that output. Otherwise, it will print an error.
func executeCommand(commandName string, commandArgs []string) string {
	tOutput, err := exec.Command(commandName, commandArgs...).Output()
	output := string(tOutput)

	if err != nil {
		return err.Error()
	}

	return output
}

// This function executes the client's shell command and returns the results.
func (s *server) SendCommand(ctx context.Context, in *pb.CommandRequest) (*pb.CommandReply, error) {

	var cmdName = in.CmdName
	var cmdArgs = in.CmdArgs
	var output = executeCommand(cmdName, cmdArgs)

	return &pb.CommandReply{Output: output}, nil
}

// Registers the client with the server as available to talk.
func (s *server) RegisterClient(ctx context.Context, in *pb.ClientInfo) (*pb.Response, error) {

	var ip = in.Ip
	var port = in.Port
	clients[ip] = port

	return &pb.Response{}, nil
}

// Unregisters the client with the server by deleting it from the global list.
func (s *server) UnRegisterClient(ctx context.Context, in *pb.ClientInfo) (*pb.Response, error) {

	var ip = in.Ip
	delete(clients, ip)

	return &pb.Response{}, nil
}

// Sends the list of currently registered IPs to the client.
func (s *server) GetClientList(ctx context.Context, in *pb.List) (*pb.ClientList, error) {

	// List of just the IPs.
	var cIps []string
	for key := range clients {
		cIps = append(cIps, key)
	}

	return &pb.ClientList{Ip: cIps}, nil
}

func (s *server) RouteChat(stream pb.Chat_RouteChatServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

//
//func (s *server) ConnectClients(ctx context.Context, in *pb.ClientIp) (*pb.Response, error) {
//
//	var ip = in.Ip
//
//}

func main() {
	// Initialize the map.
	clients = make(map[string]int32)

	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	// Initializes the gRPC server.
	s := grpc.NewServer()

	// Register the server with gRPC.
	pb.RegisterChatServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
