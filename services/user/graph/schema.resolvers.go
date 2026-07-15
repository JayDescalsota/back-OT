package graph

import (
	"context"

	"github.com/clinicmanager/services/user/graph/generated"
	"github.com/clinicmanager/services/user/graph/model"
	"github.com/clinicmanager/services/user/service"
)

type Resolver struct {
	UserService *service.UserService
}

func (r *entityResolver) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	return r.UserService.GetByID(ctx, id)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	return r.UserService.GetMe(ctx)
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	return r.UserService.GetByID(ctx, id)
}

func (r *queryResolver) MyAssignments(ctx context.Context) ([]*model.UserBranchAssignment, error) {
	return r.UserService.GetMyAssignments(ctx)
}

func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type (
	entityResolver struct{ *Resolver }
	queryResolver  struct{ *Resolver }
)
