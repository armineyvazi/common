package port

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Database interface {
	GetConnection(ctx context.Context) *gorm.DB
	Health
}

type MongoDatabase interface {
	GetConnection(ctx context.Context) *mongo.Client
	Health
}
