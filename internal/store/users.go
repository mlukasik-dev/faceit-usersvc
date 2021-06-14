package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mlukasik-dev/usersvc/pkg/deref"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func (s *Store) GetUserByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) CountUsers(ctx context.Context, filter *User) (int64, error) {
	count, err := s.users.CountDocuments(ctx, filter.filter())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) ListUsers(ctx context.Context, filter *User, p *Pagination) ([]*User, error) {
	var users []*User
	cur, err := s.users.Find(ctx, filter.filter(), p.findOpts())
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return users, err
}

func (s *Store) CreateUser(ctx context.Context, user *User, password string) (*User, error) {
	result, err := s.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := s.registerUser(sessCtx, user.Email, password); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, ErrAlreadyExists
			}
			return nil, err
		}
		result, err := s.users.InsertOne(sessCtx, user)
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrAlreadyExists
		}
		if err != nil {
			return nil, err
		}
		return s.GetUserByID(sessCtx, result.InsertedID.(primitive.ObjectID))
	})
	if err != nil {
		return nil, err
	}
	return result.(*User), err
}

func (s *Store) registerUser(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c := creds{Email: email, Password: hash}
	_, err = s.creds.ReplaceOne(ctx, bson.D{{Key: "email", Value: email}}, c, options.Replace().SetUpsert(true))
	return err
}

func (s *Store) UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	matches, err := s.matchesPassword(ctx, email, oldPassword)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !matches {
		return ErrInvalidCreds
	}
	return s.registerUser(ctx, email, newPassword)
}

func (s *Store) matchesPassword(ctx context.Context, email, password string) (bool, error) {
	var c creds
	err := s.creds.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&c)
	if err != nil {
		return false, err
	}
	if err = bcrypt.CompareHashAndPassword(c.Password, []byte(password)); err != nil {
		return false, nil
	}
	return true, nil
}

func (s *Store) UpdateUser(ctx context.Context, u *User, paths []string) (*User, error) {
	_, err := s.users.UpdateOne(ctx, bson.D{{Key: "_id", Value: u.ID}}, u.update(paths))
	if mongo.IsDuplicateKeyError(err) {
		var e mongo.WriteException
		if errors.As(err, &e) {
			for _, we := range e.WriteErrors {
				if strings.Contains(we.Message, "email_1 dup key:") {
					return nil, fmt.Errorf("user with email '%s' %w", u.Email, ErrAlreadyExists)
				} else if strings.Contains(we.Message, "nickname_1 dup key:") {
					return nil, fmt.Errorf("user with nickname '%s' %w", deref.String(u.Nickname), ErrAlreadyExists)
				}
			}
		}
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, u.ID)
}

func (s *Store) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		var u User
		err := s.users.FindOneAndDelete(sessCtx, bson.D{{Key: "_id", Value: id}}).Decode(&u)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		_, err = s.creds.DeleteOne(sessCtx, bson.D{{Key: "email", Value: u.Email}})
		return nil, err
	})
	return err
}
