package graph

import (
	"context"

	"github.com/clinicmanager/services/messaging/db"
	"github.com/clinicmanager/services/messaging/graph/generated"
)

func (r *entityResolver) FindMessageThreadByID(ctx context.Context, id string) (*db.BunMessageThread, error) {
	return r.MessagingService.GetThreadByID(ctx, id)
}

func (r *entityResolver) FindMessageByID(ctx context.Context, id string) (*db.BunMessage, error) {
	return r.MessagingService.GetMessageByID(ctx, id)
}

func (r *entityResolver) FindMessageParticipantByID(ctx context.Context, id string) (*db.BunMessageParticipant, error) {
	return r.MessagingService.Repo.FindParticipantByID(ctx, id)
}

func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
