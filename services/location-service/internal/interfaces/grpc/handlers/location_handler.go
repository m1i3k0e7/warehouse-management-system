package handlers

import (
	"context"
	"fmt"

	"warehouse/location-service/api/proto"
	"warehouse/location-service/internal/application/commands"
	"warehouse/location-service/internal/application/queries"
	"warehouse/location-service/internal/domain/entities"
)

type LocationHandler struct {
	proto.UnimplementedLocationServiceServer
	allocateSlotCmdHandler *commands.AllocateSlotCommandHandler
	findOptimalPathQueryHandler *queries.FindOptimalPathQueryHandler
	getShelfLayoutQueryHandler *queries.GetShelfLayoutQueryHandler
}

func NewLocationHandler(
	allocateSlotCmdHandler *commands.AllocateSlotCommandHandler,
	findOptimalPathQueryHandler *queries.FindOptimalPathQueryHandler,
	getShelfLayoutQueryHandler *queries.GetShelfLayoutQueryHandler,
) *LocationHandler {
	return &LocationHandler{
		allocateSlotCmdHandler: allocateSlotCmdHandler,
		findOptimalPathQueryHandler: findOptimalPathQueryHandler,
		getShelfLayoutQueryHandler: getShelfLayoutQueryHandler,
	}
}

func (h *LocationHandler) GetShelfLayout(ctx context.Context, req *proto.GetShelfLayoutRequest) (*proto.GetShelfLayoutResponse, error) {
	query := queries.GetShelfLayoutQuery{ShelfID: req.GetShelfId()}
	shelf, err := h.getShelfLayoutQueryHandler.Handle(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf layout: %w", err)
	}

	// Convert domain entity to protobuf message
	protoShelf := &proto.Shelf{
		Id:      shelf.ID,
		Name:    shelf.Name,
		Zone:    shelf.Zone,
		Rows:    shelf.Rows,
		Columns: shelf.Columns,
		Levels:  shelf.Levels,
	}
	for _, slot := range shelf.Slots {
		protoShelf.Slots = append(protoShelf.Slots, &proto.Slot{
			Id:         slot.ID,
			ShelfId:    slot.ShelfID,
			Row:        slot.Row,
			Column:     slot.Column,
			Level:      slot.Level,
			Status:     slot.Status,
			MaterialId: slot.MaterialID,
		})
	}

	return &proto.GetShelfLayoutResponse{Shelf: protoShelf}, nil
}

func (h *LocationHandler) FindOptimalPath(ctx context.Context, req *proto.FindOptimalPathRequest) (*proto.FindOptimalPathResponse, error) {
	query := queries.FindOptimalPathQuery{
		StartSlotID: req.GetStartSlotId(),
		EndSlotID:   req.GetEndSlotId(),
	}
	path, err := h.findOptimalPathQueryHandler.Handle(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find optimal path: %w", err)
	}

	return &proto.FindOptimalPathResponse{PathSlotIds: path.PathNodes}, nil
}

func (h *LocationHandler) AllocateSlot(ctx context.Context, req *proto.AllocateSlotRequest) (*proto.AllocateSlotResponse, error) {
	cmd := commands.AllocateSlotCommand{
		MaterialType: req.GetMaterialType(),
		Zone:         req.GetZone(),
	}
	slot, err := h.allocateSlotCmdHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate slot: %w", err)
	}

	return &proto.AllocateSlotResponse{SlotId: slot.ID}, nil
}
