package antipode

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	mongoClient *mongo.Client
	database    string
}

type Document struct {
	Key   string  `bson:"key"`
	Value AntiObj `bson:"value"`
}

func CreateMongoDB(host string, port string, database string) MongoDB {
	clientOptions := options.Client().ApplyURI("mongodb://" + host + ":" + port + "/?directConnection=true")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}
	return MongoDB{client,
		database,
	}
}

func (m MongoDB) write(ctx context.Context, collectionName string, key string, obj AntiObj) error {

	collection := m.mongoClient.Database(m.database).Collection(collectionName)

	mongoObj := Document{
		Key:   key,
		Value: obj,
	}

	_, err := collection.InsertOne(ctx, mongoObj)

	return err
}

func (m MongoDB) read(ctx context.Context, collection string, key string) (AntiObj, error) {

	filter := bson.D{{"key", key}}

	var result Document
	err := m.mongoClient.Database(m.database).Collection(collection).FindOne(ctx, filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return AntiObj{}, ErrNotFound
	} else if err != nil {
		return AntiObj{}, err
	}

	return result.Value, nil
}

func (m MongoDB) consume(context.Context, string, string, chan struct{}) (<-chan AntiObj, error) {
	return nil, nil
}

func (m MongoDB) barrier(ctx context.Context, lineage []WriteIdentifier, datastoreID string) error {

	for _, writeIdentifier := range lineage {
		if writeIdentifier.Dtstid == datastoreID {
			for {
				filter := bson.D{{"key", writeIdentifier.Key}}

				cursor, err := m.mongoClient.Database(m.database).Collection(writeIdentifier.TableId).Find(ctx, filter)

				if !errors.Is(err, mongo.ErrNoDocuments) && err != nil {
					return err
				} else if errors.Is(err, mongo.ErrNoDocuments) { //the version replication process is not yet completed
					continue
				} else {
					replicationDone := false
					for cursor.Next(ctx) {
						var document Document
						if err := cursor.Decode(&document); err != nil {
							cursor.Close(ctx)
							return err
						}
						if document.Value.Version == writeIdentifier.Version { //the version replication process is already completed
							replicationDone = true
							break
						}
					}
					cursor.Close(ctx)
					if replicationDone { //the version replication process is already completed
						break
					} else { //the version replication process is not yet completed
						continue
					}
				}
			}
		}
	}

	return nil
}
