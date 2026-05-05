package ports

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Database interface {
	GetConnection(ctx context.Context) *gorm.DB
	Close() error
}

type MongoDatabase interface {
	GetConnection(ctx context.Context) *mongo.Client
	Close() error
}
