package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/ot/identity-svc/graph/model"
	"github.com/ot/identity-svc/repository"
)


func (r *entityResolver) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	return r.UserService.GetByID(ctx, id)
}

func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.AuthPayload, error) {
	return r.AuthService.Register(ctx, input.Email, input.Password, input.Name)
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	return r.AuthService.Login(ctx, input.Email, input.Password)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	claims := repository.GetClaims(ctx)
	return r.UserService.GetMe(ctx, claimsID(claims))
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	return r.UserService.GetByID(ctx, id)
}

func (r *queryResolver) MyAssignments(ctx context.Context) ([]*model.UserBranchAssignment, error) {
	claims := repository.GetClaims(ctx)
	return r.UserService.GetMyAssignments(ctx, claimsID(claims))
}

func claimsID(claims *repository.UserClaims) string {
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// Entity returns EntityResolver implementation.
func (r *Resolver) Entity() EntityResolver { return &entityResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type entityResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
