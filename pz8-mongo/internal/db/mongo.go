package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeps struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongo(ctx context.Context, uri, dbName string) (*MongoDeps, error) {
	opts := options.Client().ApplyURI(uri)
	cli, err := mongo.NewClient(opts)
	if err != nil { return nil, err }

	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := cli.Connect(dialCtx); err != nil { return nil, err }

	pingCtx, cancelPing := context.WithTimeout(ctx, 3*time.Second)
	defer cancelPing()
	if err := cli.Ping(pingCtx, nil); err != nil { return nil, err }

	return &MongoDeps{Client: cli, Database: cli.Database(dbName)}, nil
}
