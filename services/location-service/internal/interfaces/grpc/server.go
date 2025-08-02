package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/m1i3k0e7/warehouse-management-system/services/location-service/api/proto"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/interfaces/grpc/handlers"
)

// Server is a gRPC server.
	type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new gRPC server.
func NewServer(port string, locationServer *handlers.LocationServer) *Server {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLocationServiceServer(s, locationServer)

	return &Server{
		grpcServer: s,
		listener:   lis,
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	log.Printf("gRPC server listening on %s", s.listener.Addr())
	return s.grpcServer.Serve(s.listener)
}

// Stop stops the gRPC server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}