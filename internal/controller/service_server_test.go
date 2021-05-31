// +build integration

package controller_test

import (
	"context"
	"testing"

	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/controller"
	"github.com/mlukasik-dev/faceit-usersvc/internal/events"
	"github.com/mlukasik-dev/faceit-usersvc/pkg/deref"
	"github.com/mlukasik-dev/faceit-usersvc/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestServiceServer_HealthCheck(t *testing.T) {
	req := &usersvcv1.HealthCheckRequest{}
	res, err := ctr.HealthCheck(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "HEALTHY", res.Status)
}

func TestServiceServer_ListUsers(t *testing.T) {
	// Use default values for pagination.
	t.Run("basic", func(t *testing.T) {
		req := &usersvcv1.ListUsersRequest{}
		res, err := ctr.ListUsers(context.Background(), req)
		require.NoError(t, err)
		assert.Len(t, res.Users, 3)
		assert.Equal(t, res.Page, int32(1))
		assert.Equal(t, res.Size, int32(15))
		assert.Equal(t, res.Total, int64(3))
	})

	t.Run("paginate", func(t *testing.T) {
		// Skip first 2 of total 3 test users.
		req := &usersvcv1.ListUsersRequest{Page: 2, Size: 2}
		res, err := ctr.ListUsers(context.Background(), req)
		require.NoError(t, err)
		assert.Len(t, res.Users, 1)
		assert.Equal(t, res.Page, int32(2))
		assert.Equal(t, res.Size, int32(2))
		assert.Equal(t, res.Total, int64(3))
	})

	t.Run("filter", func(t *testing.T) {
		// Query John Doe and Jane Doe by their surname.
		req := &usersvcv1.ListUsersRequest{Filters: &usersvcv1.User{LastName: "Doe"}}
		res, err := ctr.ListUsers(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, res.Users, 2)
		assert.Equal(t, res.Users[0].LastName, "Doe")
		assert.Equal(t, res.Users[1].LastName, "Doe")
	})

	t.Run("filter by unique field", func(t *testing.T) {
		req := &usersvcv1.ListUsersRequest{Filters: &usersvcv1.User{Email: "jan.kowalski@gmail.com"}}
		res, err := ctr.ListUsers(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, res.Users, 1)
		assert.Equal(t, res.Users[0].Email, "jan.kowalski@gmail.com")
	})

	t.Run("validate filters", func(t *testing.T) {
		reqs := []*usersvcv1.ListUsersRequest{
			{Filters: &usersvcv1.User{Email: "jan.kowalski#gmail.com"}},
			{Filters: &usersvcv1.User{FirstName: "go113"}},
			{Filters: &usersvcv1.User{LastName: "311og"}},
		}
		for _, req := range reqs {
			_, err := ctr.ListUsers(context.Background(), req)
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.InvalidArgument)
		}
	})
}

func TestServiceServer_GetUser(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		user := testData.users[1]
		req := &usersvcv1.GetUserRequest{Id: user.ID.Hex()}
		res, err := ctr.GetUser(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, res.Id, user.ID.Hex())
		assert.Equal(t, res.FirstName, user.FirstName)
		assert.Equal(t, res.LastName, user.LastName)
		assert.Equal(t, res.Nickname, deref.String(user.Nickname))
		assert.Equal(t, res.Email, user.Email)
		assert.Equal(t, res.Country, user.Country)
	})

	t.Run("not existing", func(t *testing.T) {
		req := &usersvcv1.GetUserRequest{Id: primitive.NewObjectID().Hex()}
		_, err := ctr.GetUser(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, status.Convert(err).Code(), codes.NotFound)
	})
}

func TestServiceServer_CreateUser(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		e := &events.Mock{}
		e.On("Publish", events.CreateUserEvent, "<id>").Return()
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			user := &usersvcv1.User{FirstName: "Mark", LastName: "Brown", Nickname: "mb", Email: "mark.brown@gmail.com", Country: "US"}
			req := &usersvcv1.CreateUserRequest{User: user, Password: ""}
			res, err := ctr.CreateUser(ctx, req)
			e.AssertExpectations(t)
			require.NoError(t, err)
			assert.NotEqual(t, res.Id, "")
			assert.Equal(t, res.FirstName, user.FirstName)
			assert.Equal(t, res.LastName, user.LastName)
			assert.Equal(t, res.Nickname, user.Nickname)
			assert.Equal(t, res.Email, user.Email)
			assert.Equal(t, res.Country, user.Country)
		})
	})

	t.Run("validate", func(t *testing.T) {
		e := &events.Mock{}
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			user := &usersvcv1.User{FirstName: "Mark", LastName: "Brown", Nickname: "#-#", Email: "mark.brown@gmail.com", Country: "US"}
			req := &usersvcv1.CreateUserRequest{User: user, Password: ""}
			_, err := ctr.CreateUser(ctx, req)
			e.AssertNotCalled(t, "Publish")
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.InvalidArgument)
		})
	})

	t.Run("already exists", func(t *testing.T) {
		e := &events.Mock{}
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			user := &usersvcv1.User{FirstName: "John", LastName: "Doe", Email: "john.doe@gmail.com", Country: "UK"}
			req := &usersvcv1.CreateUserRequest{User: user, Password: ""}
			_, err := ctr.CreateUser(ctx, req)
			e.AssertNotCalled(t, "Publish")
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.AlreadyExists)
		})
	})
}

func TestServiceServer_UpdatePassword(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			req := &usersvcv1.UpdatePasswordRequest{Email: "jane.doe@gmail.com", OldPassword: "123456", NewPassword: "654321"}
			_, err := ctr.UpdatePassword(ctx, req)
			require.NoError(t, err)

			req = &usersvcv1.UpdatePasswordRequest{Email: "jane.doe@gmail.com", OldPassword: "123456", NewPassword: "654321"}
			_, err = ctr.UpdatePassword(ctx, req)
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.PermissionDenied)

			req = &usersvcv1.UpdatePasswordRequest{Email: "jane.doe@gmail.com", OldPassword: "654321", NewPassword: "123456"}
			_, err = ctr.UpdatePassword(ctx, req)
			require.NoError(t, err)
		})
	})

	t.Run("invalid creds", func(t *testing.T) {
		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			req := &usersvcv1.UpdatePasswordRequest{Email: "jane.doe@gmail.com", OldPassword: "", NewPassword: "654321"}
			_, err := ctr.UpdatePassword(ctx, req)
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.PermissionDenied)
		})
	})
}

func TestServiceServer_UpdateUser(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := testData.users[0]
		e := &events.Mock{}
		e.On("Publish", events.UpdateUserEvent, user.ID).Return()
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			pbUser := &usersvcv1.User{Id: user.ID.Hex(), Country: "PL"}
			um, err := fieldmaskpb.New(pbUser, "country")
			require.NoError(t, err)
			req := &usersvcv1.UpdateUserRequest{User: pbUser, UpdateMask: um}
			res, err := ctr.UpdateUser(ctx, req)
			e.AssertExpectations(t)
			require.NoError(t, err)
			assert.Equal(t, res.Id, user.ID.Hex())
			assert.Equal(t, res.FirstName, user.FirstName)
			assert.Equal(t, res.LastName, user.LastName)
			assert.Equal(t, res.Nickname, deref.String(user.Nickname))
			assert.Equal(t, res.Email, user.Email)
			assert.Equal(t, res.Country, "PL")
		})
	})

	t.Run("conflict", func(t *testing.T) {
		// Try to change user's email to email of an existing user.
		user := testData.users[0]
		e := &events.Mock{}
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			pbUser := &usersvcv1.User{Id: user.ID.Hex(), Email: "jan.kowalski@gmail.com"}
			um, err := fieldmaskpb.New(pbUser, "email")
			require.NoError(t, err)
			req := &usersvcv1.UpdateUserRequest{User: pbUser, UpdateMask: um}
			_, err = ctr.UpdateUser(ctx, req)
			e.AssertNotCalled(t, "Publish")
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.AlreadyExists)
		})
	})
}

func TestServiceServer_DeleteUser(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		id := testData.users[1].ID
		e := &events.Mock{}
		e.On("Publish", events.DeleteUserEvent, id).Return()
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			req := &usersvcv1.DeleteUserRequest{Id: id.Hex()}
			_, err := ctr.DeleteUser(ctx, req)
			e.AssertExpectations(t)
			require.NoError(t, err)
			{
				req := &usersvcv1.GetUserRequest{Id: id.Hex()}
				_, err = ctr.GetUser(ctx, req)
				require.Error(t, err)
				assert.Equal(t, status.Convert(err).Code(), codes.NotFound)
			}
		})
	})

	t.Run("not existing", func(t *testing.T) {
		e := &events.Mock{}
		ctr := controller.New(s, l, e)

		testutils.WithAbortedTransaction(context.Background(), s.Client(), func(ctx context.Context) {
			req := &usersvcv1.DeleteUserRequest{Id: primitive.NewObjectID().Hex()}
			_, err := ctr.DeleteUser(ctx, req)
			e.AssertNotCalled(t, "Publish")
			require.Error(t, err)
			assert.Equal(t, status.Convert(err).Code(), codes.NotFound)
		})
	})
}
