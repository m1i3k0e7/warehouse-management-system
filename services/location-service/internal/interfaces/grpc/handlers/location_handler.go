package handlers

import (
	"context"

	pb "github.com/m1i3k0e7/warehouse-management-system/services/location-service/api/proto"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/application/commands"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/application/queries"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/repositories"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LocationServer is the implementation of the gRPC LocationService.
	ype LocationServer struct {
	pb.UnimplementedLocationServiceServer

	// Repositories
	shelfRepo  repositories.ShelfRepository
	layoutRepo repositories.LayoutRepository

	// Domain Services
	pathfinder        *services.PathfindingService
	allocationService *services.AllocationService
}

// NewLocationServer creates a new LocationServer.
func NewLocationServer(shelfRepo repositories.ShelfRepository, layoutRepo repositories.LayoutRepository) *LocationServer {
	return &LocationServer{
		shelfRepo:         shelfRepo,
		layoutRepo:        layoutRepo,
		pathfinder:        services.NewPathfindingService(),
		allocationService: services.NewAllocationService(layoutRepo),
	}
}

func (s *LocationServer) GetShelfLayout(ctx context.Context, req *pb.GetShelfLayoutRequest) (*pb.ShelfLayoutResponse, error) {
	q := queries.NewGetShelfLayoutQueryHandler(s.shelfRepo)
	shelf, err := q.Handle(ctx, req.ShelfId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get shelf layout: %v", err)
	}
	if shelf == nil {
		return nil, status.Errorf(codes.NotFound, "shelf with id %s not found", req.ShelfId)
	}

	return &pb.ShelfLayoutResponse{Shelf: toProtoShelf(shelf)}, nil
}

func (s *LocationServer) FindOptimalPath(ctx context.Context, req *pb.FindOptimalPathRequest) (*pb.FindOptimalPathResponse, error) {
	q := queries.NewFindOptimalPathQueryHandler(s.pathfinder)
	path, err := q.Handle(ctx, fromProtoPoint(req.StartPoint), fromProtoPoint(req.EndPoint))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find optimal path: %v", err)
	}

	return &pb.FindOptimalPathResponse{Path: toProtoPath(path.Points), Distance: path.Distance}, nil
}

func (s *LocationServer) SuggestPlacement(ctx context.Context, req *pb.SuggestPlacementRequest) (*pb.SuggestPlacementResponse, error) {
	cmd := commands.NewAllocateSlotCommandHandler(s.allocationService, s.shelfRepo)
	shelf, slot, err := cmd.Handle(ctx, req.MaterialType, req.ZoneId, "") // materialID is empty because we are just suggesting
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to suggest placement: %v", err)
	}
	if shelf == nil || slot == nil {
		return nil, status.Errorf(codes.NotFound, "no available slot found in zone %s", req.ZoneId)
	}

	return &pb.SuggestPlacementResponse{ShelfId: shelf.ID, SlotId: slot.ID}, nil
}

// --- Converters ---

func toProtoShelf(shelf *entities.Shelf) *pb.Shelf {
	// ... implementation ...
	return &pb.Shelf{}
}

func fromProtoPoint(p *pb.Point) entities.Point {
	// ... implementation ...
	return entities.Point{}
}

func toProtoPath(points []entities.Point) []*pb.Point {
	// ... implementation ...
	return []*pb.Point{}
}
