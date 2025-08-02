package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/repositories"
)

// MongoRepository is a MongoDB implementation of the repositories.
type MongoRepository struct {
	client   *mongo.Client
	database string
}

// NewMongoRepository creates a new MongoRepository.
func NewMongoRepository(ctx context.Context, uri, database string) (*MongoRepository, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoRepository{
		client:   client,
		database: database,
	}, nil
}

// Disconnect disconnects the client from MongoDB.
func (r *MongoRepository) Disconnect(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

// --- ShelfRepository Implementation ---

func (r *MongoRepository) shelves() *mongo.Collection {
	return r.client.Database(r.database).Collection("shelves")
}

func (r *MongoRepository) FindByID(ctx context.Context, id string) (*entities.Shelf, error) {
	var shelf entities.Shelf
	err := r.shelves().FindOne(ctx, bson.M{"id": id}).Decode(&shelf)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return &shelf, nil
}

func (r *MongoRepository) Save(ctx context.Context, shelf *entities.Shelf) error {
	opts := options.Replace().SetUpsert(true)
	_, err := r.shelves().ReplaceOne(ctx, bson.M{"id": shelf.ID}, shelf, opts)
	return err
}

func (r *MongoRepository) UpdateSlotStatus(ctx context.Context, shelfID string, slotID string, status entities.SlotStatus, materialID string) error {
	filter := bson.M{"id": shelfID, "slots.id": slotID}
	update := bson.M{"$set": bson.M{
		"slots.$.status":     status,
		"slots.$.materialid": materialID,
	}}
	_, err := r.shelves().UpdateOne(ctx, filter, update)
	return err
}

// --- LayoutRepository Implementation ---

func (r *MongoRepository) zones() *mongo.Collection {
	return r.client.Database(r.database).Collection("zones")
}

func (r *MongoRepository) FindZoneByID(ctx context.Context, id string) (*entities.Zone, error) {
	var zone entities.Zone
	err := r.zones().FindOne(ctx, bson.M{"id": id}).Decode(&zone)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &zone, nil
}

func (r *MongoRepository) SaveZone(ctx context.Context, zone *entities.Zone) error {
	opts := options.Replace().SetUpsert(true)
	_, err := r.zones().ReplaceOne(ctx, bson.M{"id": zone.ID}, zone, opts)
	return err
}

func (r *MongoRepository) FindAllShelvesInZone(ctx context.Context, zoneID string) ([]*entities.Shelf, error) {
	cursor, err := r.shelves().Find(ctx, bson.M{"zoneid": zoneID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shelves []*entities.Shelf
	if err = cursor.All(ctx, &shelves); err != nil {
		return nil, err
	}
	return shelves, nil
}

// Ensure MongoRepository implements the interfaces
var _ repositories.ShelfRepository = (*MongoRepository)(nil)
var _ repositories.LayoutRepository = (*MongoRepository)(nil)