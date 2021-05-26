package controller

import (
	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/store"
	"github.com/mlukasik-dev/faceit-usersvc/pkg/deref"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userToPb(u *store.User) *usersvcv1.User {
	return &usersvcv1.User{
		Id:        u.ID.Hex(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  deref.String(u.Nickname),
		Email:     u.Email,
		Country:   u.Country,
	}
}

func pbToUser(pb *usersvcv1.User) *store.User {
	id, _ := primitive.ObjectIDFromHex(pb.Id)
	return &store.User{
		ID:        id,
		FirstName: pb.FirstName,
		LastName:  pb.LastName,
		Nickname:  &pb.Nickname,
		Email:     pb.Email,
		Country:   pb.Country,
	}
}
