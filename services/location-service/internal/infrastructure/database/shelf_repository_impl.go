package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/repositories"
)

type mongoShelfRepository struct {
	collection *mongo.Collection
}

func NewMongoShelfRepository(db *MongoDB) repositories.ShelfRepository {
	return &mongoShelfRepository{
		collection: db.Database.Collection("shelves"),
	}
}

func (r *mongoShelfRepository) CreateShelf(ctx context.Context, shelf *entities.Shelf) error {
	shelf.CreatedAt = time.Now()
	shelf.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, shelf)
	return err
}

func (r *mongoShelfRepository) GetShelfByID(ctx context.Context, id string) (*entities.Shelf, error) {
	var shelf entities.Shelf
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&shelf)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("shelf not found")
	}
	return &shelf, err
}

func (r *mongoShelfRepository) GetAllShelves(ctx context.Context) ([]entities.Shelf, error) {
	var shelves []entities.Shelf
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &shelves); err != nil {
		return nil, err
	}
	return shelves, nil
}

func (r *mongoShelfRepository) GetShelvesByZone(ctx context.Context, zone string) ([]entities.Shelf, error) {
	var shelves []entities.Shelf
	cursor, err := r.collection.Find(ctx, bson.M{"zone": zone})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &shelves); err != nil {
		return nil, err
	}
	return shelves, nil
}

func (r *mongoShelfRepository) UpdateShelf(ctx context.Context, shelf *entities.Shelf) error {
	shelf.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": shelf.ID}, shelf)
	return err
}

func (r *mongoShelfRepository) DeleteShelf(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoShelfRepository) CreateSlot(ctx context.Context, slot *entities.Slot) error {
	// Find the shelf and append the new slot
	filter := bson.M{"_id": slot.ShelfID}
	update := bson.M{"$push": bson.M{"slots": slot}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *mongoShelfRepository) GetSlotByID(ctx context.Context, id string) (*entities.Slot, error) {
	// This requires iterating through all shelves or using an aggregation pipeline
	// For simplicity, let's assume slot IDs are unique across all shelves and we can find it.
	// A more efficient way would be to have slots as a separate collection or use a specific index.
	var shelf entities.Shelf
	err := r.collection.FindOne(ctx, bson.M{"slots.id": id}).Decode(&shelf)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("slot not found")
	}
	if err != nil {
		return nil, err
	}

	for _, slot := range shelf.Slots {
		if slot.ID == id {
			return &slot, nil
		}
	}
	return nil, fmt.Errorf("slot not found within the shelf")
}

func (r *mongoShelfRepository) UpdateSlot(ctx context.Context, slot *entities.Slot) error {
	// This is complex with embedded documents. Requires finding the shelf and then updating the specific slot within the array.
	// A better approach for frequent slot updates might be to store slots in a separate collection.
	filter := bson.M{"_id": slot.ShelfID, "slots.id": slot.ID}
	update := bson.M{"$set": bson.M{
		"slots.$.status":     slot.Status,
		"slots.$.material_id": slot.MaterialID,
		// Update other fields as needed
	}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *mongoShelfRepository) DeleteSlot(ctx context.Context, id string) error {
	// This is also complex for embedded documents. Requires pulling the element from the array.
	// You'd need to find the shelf first, then pull the slot.
	// For simplicity, this implementation assumes you know the shelf ID or will find it.
	// A more robust solution would involve a separate slots collection.
	return fmt.Errorf("delete slot not implemented for embedded documents")
}

func (r *mongoShelfRepository) GetSlotsByShelfID(ctx context.Context, shelfID string) ([]entities.Slot, error) {
	shelf, err := r.GetShelfByID(ctx, shelfID)
	if err != nil {
		return nil, err
	}
	return shelf.Slots, nil
}
