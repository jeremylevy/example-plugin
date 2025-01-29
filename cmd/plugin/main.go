// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/newco/plugin-proto/proto"
	"google.golang.org/grpc"
)

func findAvailablePort() (int, error) {
	// Let the system assign a random available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	// Get the actual address being used
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

type server struct {
	pb.UnimplementedPluginServer
}

func (s *server) GetMetadata(ctx context.Context, req *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	return &pb.MetadataResponse{
		Name:        "example-plugin",
		Version:     "1.0.0",
		Description: "An example plugin implementation",
	}, nil
}

func (s *server) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	log.Printf("Executing action: %s with data: %s", req.Action, string(req.Data))
	return &pb.ExecuteResponse{
		Success: true,
		Result:  req.Data,
	}, nil
}

func main() {
	if os.Getenv("PLUGIN_SERVER") != "true" {
		log.Fatal("This binary is a plugin and should not be run directly")
	}

	// Find an available port
	port, err := findAvailablePort()
	if err != nil {
		log.Fatalf("failed to find available port: %v", err)
	}

	// Print the port number to stdout in a parseable format
	// This is how we communicate the port back to the CLI
	fmt.Printf("PORT=%d\n", port)

	// Start the gRPC server on the chosen port
	address := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPluginServer(s, &server{})

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		s.GracefulStop()
	}()

	log.Printf("Plugin server starting on port %d", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
