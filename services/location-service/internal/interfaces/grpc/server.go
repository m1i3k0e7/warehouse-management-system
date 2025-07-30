package grpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"warehouse/location-service/api/proto"
	"warehouse/location-service/internal/interfaces/grpc/handlers"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func NewServer(port string, locationHandler *handlers.LocationHandler) *Server {
	grpcServer := grpc.NewServer()
	proto.RegisterLocationServiceServer(grpcServer, locationHandler)
	reflection.Register(grpcServer) // Enable gRPC reflection for tools like grpcurl

	return &Server{
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on port %s", s.port)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}
