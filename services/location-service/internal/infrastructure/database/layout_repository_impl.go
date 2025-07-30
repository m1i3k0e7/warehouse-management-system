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

type mongoLayoutRepository struct {
	zoneCollection *mongo.Collection
	pathCollection *mongo.Collection
}

func NewMongoLayoutRepository(db *MongoDB) repositories.LayoutRepository {
	return &mongoLayoutRepository{
		zoneCollection: db.Database.Collection("zones"),
		pathCollection: db.Database.Collection("paths"),
	}
}

func (r *mongoLayoutRepository) CreateZone(ctx context.Context, zone *entities.Zone) error {
	zone.CreatedAt = time.Now()
	zone.UpdatedAt = time.Now()
	_, err := r.zoneCollection.InsertOne(ctx, zone)
	return err
}

func (r *mongoLayoutRepository) GetZoneByID(ctx context.Context, id string) (*entities.Zone, error) {
	var zone entities.Zone
	err := r.zoneCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&zone)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("zone not found")
	}
	return &zone, err
}

func (r *mongoLayoutRepository) GetAllZones(ctx context.Context) ([]entities.Zone, error) {
	var zones []entities.Zone
	cursor, err := r.zoneCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *mongoLayoutRepository) UpdateZone(ctx context.Context, zone *entities.Zone) error {
	zone.UpdatedAt = time.Now()
	_, err := r.zoneCollection.ReplaceOne(ctx, bson.M{"_id": zone.ID}, zone)
	return err
}

func (r *mongoLayoutRepository) DeleteZone(ctx context.Context, id string) error {
	_, err := r.zoneCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoLayoutRepository) CreatePath(ctx context.Context, path *entities.Path) error {
	path.CreatedAt = time.Now()
	_, err := r.pathCollection.InsertOne(ctx, path)
	return err
}

func (r *mongoLayoutRepository) GetPathByID(ctx context.Context, id string) (*entities.Path, error) {
	var path entities.Path
	err := r.pathCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&path)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("path not found")
	}
	return &path, err
}

func (r *mongoLayoutRepository) GetPathsBySlots(ctx context.Context, startSlotID, endSlotID string) ([]entities.Path, error) {
	var paths []entities.Path
	filter := bson.M{
		"start_slot": startSlotID,
		"end_slot":   endSlotID,
	}
	cursor, err := r.pathCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &paths); err != nil {
		return nil, err
	}
	return paths, nil
}

func (r *mongoLayoutRepository) UpdatePath(ctx context.Context, path *entities.Path) error {
	_, err := r.pathCollection.ReplaceOne(ctx, bson.M{"_id": path.ID}, path)
	return err
}

func (r *mongoLayoutRepository) DeletePath(ctx context.Context, id string) error {
	_, err := r.pathCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
