package controller

import (
	"context"
	"errors"

	"github.com/gookit/validate"
	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/events"
	"github.com/mlukasik-dev/faceit-usersvc/internal/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Ctr struct {
	store  *store.Store
	logger *zap.Logger
	events *events.Client
}

func New(s *store.Store, l *zap.Logger, e *events.Client) usersvcv1.ServiceServer {
	return &Ctr{s, l, e}
}

func (ctr *Ctr) ListUsers(ctx context.Context, req *usersvcv1.ListUsersRequest) (*usersvcv1.ListUsersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	if req.Page < 0 {
		return nil, status.Error(codes.InvalidArgument, "req.page should be a positive integer")
	}
	if req.Size < 0 {
		return nil, status.Error(codes.InvalidArgument, "req.size should be a positive integer")
	}
	if req.Page == 0 {
		req.Page = 1 // default page.
	}
	if req.Size == 0 {
		req.Size = 15 // default size.
	}
	if req.Filters == nil {
		req.Filters = &usersvcv1.User{}
	}
	filter := pbToUser(req.Filters)
	if err := filter.Validate(store.FilterValidationKind); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.One())
	}

	count, err := ctr.store.CountUsers(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	users, err := ctr.store.ListUsers(ctx, filter, &store.Pagination{Page: uint(req.Page), Size: uint(req.Size)})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &usersvcv1.ListUsersResponse{
		Page:  req.Page,
		Size:  req.Size,
		Total: count,
	}
	for _, u := range users {
		resp.Users = append(resp.Users, userToPb(u))
	}
	return resp, nil
}

func (ctr *Ctr) GetUser(ctx context.Context, req *usersvcv1.GetUserRequest) (*usersvcv1.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	u, err := ctr.store.GetUserByID(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userToPb(u), nil
}

func (ctr *Ctr) CreateUser(ctx context.Context, req *usersvcv1.CreateUserRequest) (*usersvcv1.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "req.user should not be <nil>")
	}
	u := pbToUser(req.User)
	if err := u.Validate(store.CreateValidationKind); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.One())
	}

	u, err := ctr.store.CreateUser(ctx, u, req.Password)
	if errors.Is(err, store.ErrAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ctr.events.Publish(events.CreateUserEvent, u.ID)
	return userToPb(u), nil
}

func (ctr *Ctr) UpdatePassword(ctx context.Context, req *usersvcv1.UpdatePasswordRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	if !validate.IsEmail(req.Email) {
		return nil, status.Error(codes.NotFound, "invalid email")
	}

	err := ctr.store.UpdatePassword(ctx, req.Email, req.OldPassword, req.NewPassword)
	if errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, store.ErrInvalidCreds) {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (ctr *Ctr) UpdateUser(ctx context.Context, req *usersvcv1.UpdateUserRequest) (*usersvcv1.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "req.user should not be <nil>")
	}
	if req.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "req.update_mask should not be <nil>")
	}
	if !req.UpdateMask.IsValid(req.User) || len(req.UpdateMask.Paths) == 0 || contains(req.UpdateMask.Paths, "id") {
		return nil, status.Error(codes.InvalidArgument, "invalid update_mask")
	}
	u, err := pbToUser(req.User).SetID(req.User.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := u.Validate(store.UpdateValidationKind); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.One())
	}
	u, err = ctr.store.UpdateUser(ctx, u, req.UpdateMask.Paths)
	if errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ctr.events.Publish(events.UpdateUserEvent, u.ID)
	return userToPb(u), nil
}

func (ctr *Ctr) DeleteUser(ctx context.Context, req *usersvcv1.DeleteUserRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req should not be <nil>")
	}
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = ctr.store.DeleteUser(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ctr.events.Publish(events.DeleteUserEvent, id)
	return &emptypb.Empty{}, nil
}

func (ctr *Ctr) HealthCheck(ctx context.Context, _ *usersvcv1.HealthCheckRequest) (*usersvcv1.HealthCheckResponse, error) {
	if err := ctr.store.Ping(ctx); err != nil {
		ctr.logger.Error("mongodb ping failed", zap.String("error", err.Error()))
		return nil, status.Error(codes.Unavailable, "NOT_HEALTHY")
	}
	return &usersvcv1.HealthCheckResponse{
		Status: "HEALTHY",
	}, nil
}
