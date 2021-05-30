package store

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is an object reused as CreateUserDTO, UpdateUserDTO and filter for ListUsers,
// and as a business object representing user resource.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" validate:"-"`
	FirstName string             `bson:"firstName" validate:"alpha|required_if:validationKind,create"`
	LastName  string             `bson:"lastName" validate:"alpha|required_if:validationKind,create"`
	Nickname  *string            `bson:"nickname" validate:"alphaNum"`
	Email     string             `bson:"email" validate:"email|required_if:validationKind,create"`
	Country   string             `bson:"country" validate:"required_if:validationKind,create"`
}

// SetID parses hex id and sets it on user object.
func (u *User) SetID(hex string) (*User, error) {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return u, err
	}
	u.ID = id
	return u, nil
}

// filter creates mongodb a document containing query operators.
// Ignores ID field.
func (filter *User) filter() bson.D {
	d := bson.D{}
	if filter != nil {
		if filter.FirstName != "" {
			d = append(d, bson.E{Key: "firstName", Value: filter.FirstName})
		}
		if filter.LastName != "" {
			d = append(d, bson.E{Key: "lastName", Value: filter.LastName})
		}
		if filter.Nickname != nil && *filter.Nickname != "" {
			d = append(d, bson.E{Key: "nickname", Value: *filter.Nickname})
		}
		if filter.Email != "" {
			d = append(d, bson.E{Key: "email", Value: filter.Email})
		}
		if filter.Country != "" {
			d = append(d, bson.E{Key: "country", Value: filter.Country})
		}
	}
	return d
}

// filter creates a mongodb document containing update operators.
// Ignores ID field.
func (u *User) update(paths []string) bson.D {
	var d bson.D
	for _, path := range paths {
		switch path {
		case "first_name":
			d = append(d, bson.E{Key: "$set", Value: bson.D{{Key: "firstName", Value: u.FirstName}}})
		case "last_name":
			d = append(d, bson.E{Key: "$set", Value: bson.D{{Key: "lastName", Value: u.LastName}}})
		case "nickname":
			d = append(d, bson.E{Key: "$set", Value: bson.D{{Key: "nickname", Value: u.Nickname}}})
		case "email":
			d = append(d, bson.E{Key: "$set", Value: bson.D{{Key: "email", Value: u.Email}}})
		case "country":
			d = append(d, bson.E{Key: "$set", Value: bson.D{{Key: "country", Value: u.Country}}})
		}
	}
	return d
}

type creds struct {
	Email    string `bson:"email"`
	Password []byte `bson:"password"`
}
