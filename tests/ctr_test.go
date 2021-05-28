package tests

import (
	"context"
	"testing"

	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestHealthCheck(t *testing.T) {
	req := &usersvcv1.HealthCheckRequest{}
	res, err := ctr.HealthCheck(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "HEALTHY" {
		t.Fail()
	}
}

func TestListUsers(t *testing.T) {
	req := &usersvcv1.ListUsersRequest{}
	res, err := ctr.ListUsers(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Users) != len(testData.users) {
		t.Fail()
	}
	if res.Page != 1 || req.Size != 15 {
		t.Fail()
	}
}

func TestGetUser(t *testing.T) {
	usersReq := &usersvcv1.ListUsersRequest{Filters: &usersvcv1.User{FirstName: testData.users[1].FirstName}}
	usersRes, err := ctr.ListUsers(context.Background(), usersReq)
	if err != nil {
		t.Fatal(err)
	}
	id := usersRes.Users[0].Id

	req := &usersvcv1.GetUserRequest{Id: id}
	res, err := ctr.GetUser(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.FirstName != testData.users[1].FirstName {
		t.Fail()
	}
}

func TestCreateUser(t *testing.T) {
	user := &usersvcv1.User{FirstName: "Mark", LastName: "Brown", Nickname: "mb", Email: "mark.brown@gmail.com", Country: "US"}
	req := &usersvcv1.CreateUserRequest{User: user, Password: ""}
	res, err := ctr.CreateUser(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.FirstName != user.FirstName ||
		res.LastName != user.LastName ||
		res.Nickname != user.Nickname ||
		res.Email != user.Email ||
		res.Country != user.Country {
		t.Fail()
	}
}

func TestUpdateUser(t *testing.T) {
	user := &usersvcv1.User{Id: testData.users[1].ID.Hex(), Country: "PL"}
	um, err := fieldmaskpb.New(user, "country")
	if err != nil {
		t.Fatal(err)
	}
	req := &usersvcv1.UpdateUserRequest{User: user, UpdateMask: um}
	res, err := ctr.UpdateUser(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Country != "PL" {
		t.Fail()
	}
}

func TestDeleteUser(t *testing.T) {
	req := &usersvcv1.DeleteUserRequest{Id: testData.users[1].ID.Hex()}
	_, err := ctr.DeleteUser(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	getReq := &usersvcv1.GetUserRequest{Id: testData.users[1].ID.Hex()}
	_, err = ctr.GetUser(context.Background(), getReq)
	if status.Convert(err).Code() != codes.NotFound {
		t.Fail()
	}
}
