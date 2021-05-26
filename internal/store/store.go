package store

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidCreds  = errors.New("invalid credentials")
)

type Store struct {
	client *mongo.Client
	users  *mongo.Collection
	creds  *mongo.Collection
}

func New(client *mongo.Client) *Store {
	db := client.Database("usersvcdb")
	users := db.Collection("users")
	creds := db.Collection("creds")
	return &Store{client, users, creds}
}

func (s *Store) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := s.client.StartSession()
	defer session.EndSession(ctx)
	if err != nil {
		return nil, err
	}
	return session.WithTransaction(ctx, fn)
}

func (s *Store) CreateIndexes(ctx context.Context) error {
	// Unique index for email field on users collection.
	_, err := s.users.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return err
	}

	// Unique index for nickname field on users collection,
	// but nulls are allowed.
	_, err = s.users.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "nickname", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "nickname", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
	)
	if err != nil {
		return err
	}

	// Unique index for email field on creds collection.
	_, err = s.creds.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Ping pings db with 3 seconds timeout.
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.client.Ping(ctx, nil)
}
