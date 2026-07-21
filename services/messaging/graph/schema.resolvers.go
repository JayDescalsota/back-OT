package graph

import (
	"context"
	"time"

	"github.com/clinicmanager/services/messaging/db"
	"github.com/clinicmanager/services/messaging/graph/generated"
	"github.com/clinicmanager/services/messaging/graph/model"
)

func (r *messageResolver) CreatedAt(ctx context.Context, obj *db.BunMessage) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *messageResolver) UpdatedAt(ctx context.Context, obj *db.BunMessage) (string, error) {
	return obj.UpdatedAt.Format(time.RFC3339), nil
}

func (r *messageParticipantResolver) LastReadAt(ctx context.Context, obj *db.BunMessageParticipant) (string, error) {
	return obj.LastReadAt.Format(time.RFC3339), nil
}

func (r *messageThreadResolver) Messages(ctx context.Context, obj *db.BunMessageThread) ([]*db.BunMessage, error) {
	return r.MessagingService.GetMessagesByThread(ctx, obj.ID)
}

func (r *messageThreadResolver) Participants(ctx context.Context, obj *db.BunMessageThread) ([]*db.BunMessageParticipant, error) {
	return r.MessagingService.GetParticipantsByThread(ctx, obj.ID)
}

func (r *messageThreadResolver) CreatedAt(ctx context.Context, obj *db.BunMessageThread) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *messageThreadResolver) UpdatedAt(ctx context.Context, obj *db.BunMessageThread) (string, error) {
	return obj.UpdatedAt.Format(time.RFC3339), nil
}

func (r *mutationResolver) CreateThread(ctx context.Context, input model.ThreadInput) (*db.BunMessageThread, error) {
	return r.MessagingService.CreateThread(ctx, input)
}

func (r *mutationResolver) UpdateThread(ctx context.Context, id string, input model.ThreadUpdateInput) (*db.BunMessageThread, error) {
	return r.MessagingService.UpdateThread(ctx, id, input)
}

func (r *mutationResolver) DeleteThread(ctx context.Context, id string) (bool, error) {
	return r.MessagingService.DeleteThread(ctx, id)
}

func (r *mutationResolver) SendMessage(ctx context.Context, input model.MessageInput) (*db.BunMessage, error) {
	return r.MessagingService.SendMessage(ctx, input)
}

func (r *mutationResolver) DeleteMessage(ctx context.Context, id string) (bool, error) {
	return r.MessagingService.DeleteMessage(ctx, id)
}

func (r *mutationResolver) AddParticipant(ctx context.Context, input model.ParticipantInput) (*db.BunMessageParticipant, error) {
	return r.MessagingService.AddParticipant(ctx, input)
}

func (r *mutationResolver) RemoveParticipant(ctx context.Context, id string) (bool, error) {
	return r.MessagingService.RemoveParticipant(ctx, id)
}

func (r *queryResolver) Thread(ctx context.Context, id string) (*db.BunMessageThread, error) {
	return r.MessagingService.GetThreadByID(ctx, id)
}

func (r *queryResolver) ThreadsByParticipant(ctx context.Context, participantID string) ([]*db.BunMessageThread, error) {
	return r.MessagingService.GetThreadsByParticipant(ctx, participantID)
}

func (r *queryResolver) Messages(ctx context.Context, threadID string) ([]*db.BunMessage, error) {
	return r.MessagingService.GetMessagesByThread(ctx, threadID)
}

func (r *queryResolver) Participants(ctx context.Context, threadID string) ([]*db.BunMessageParticipant, error) {
	return r.MessagingService.GetParticipantsByThread(ctx, threadID)
}

func (r *Resolver) Message() generated.MessageResolver { return &messageResolver{r} }
func (r *Resolver) MessageParticipant() generated.MessageParticipantResolver {
	return &messageParticipantResolver{r}
}
func (r *Resolver) MessageThread() generated.MessageThreadResolver { return &messageThreadResolver{r} }
func (r *Resolver) Mutation() generated.MutationResolver           { return &mutationResolver{r} }
func (r *Resolver) Query() generated.QueryResolver                 { return &queryResolver{r} }

type (
	messageResolver            struct{ *Resolver }
	messageParticipantResolver struct{ *Resolver }
	messageThreadResolver      struct{ *Resolver }
	mutationResolver           struct{ *Resolver }
	queryResolver              struct{ *Resolver }
)
