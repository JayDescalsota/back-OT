package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clinicmanager/services/messaging/db"
	"github.com/clinicmanager/services/messaging/graph/model"
	"github.com/clinicmanager/services/messaging/repository"
	"github.com/clinicmanager/shared/cache"
	sharedCtx "github.com/clinicmanager/shared/context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MessagingService struct {
	Repo      *repository.MessagingRepository
	Cache     *redis.Client
	Encryptor *Encryptor
}

func NewMessagingService(repo *repository.MessagingRepository, cacheClient *redis.Client, encryptor *Encryptor) *MessagingService {
	return &MessagingService{Repo: repo, Cache: cacheClient, Encryptor: encryptor}
}

func (s *MessagingService) cacheGet(ctx context.Context, key string, dest interface{}) (bool, error) {
	if s.Cache == nil {
		return false, nil
	}
	val, err := s.Cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *MessagingService) cacheSet(ctx context.Context, key string, val interface{}) error {
	if s.Cache == nil {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.Cache.Set(ctx, key, data, cache.DefaultTTL).Err()
}

func (s *MessagingService) cacheDel(ctx context.Context, keys ...string) {
	if s.Cache == nil {
		return
	}
	s.Cache.Del(ctx, keys...)
}

// Thread

func (s *MessagingService) GetThreadByID(ctx context.Context, id string) (*db.BunMessageThread, error) {
	ck := cache.Key("messaging", "thread", id)
	var cached db.BunMessageThread
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.Repo.FindThreadByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *MessagingService) GetThreadsByParticipant(ctx context.Context, participantID *string) ([]*db.BunMessageThread, error) {
	tctx := sharedCtx.FromContext(ctx)
	targetID := tctx.UserID
	if participantID != nil && *participantID != "" {
		targetID = *participantID
	}
	return s.Repo.FindThreadsByParticipant(ctx, targetID)
}

func (s *MessagingService) CreateThread(ctx context.Context, input model.ThreadInput) (*db.BunMessageThread, error) {
	tctx := sharedCtx.FromContext(ctx)
	if tctx.UserID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}
	if tctx.TenantID == "" {
		return nil, fmt.Errorf("tenant not identified")
	}

	now := time.Now()

	m := &db.BunMessageThread{
		ID:            uuid.NewString(),
		TenantID:      tctx.TenantID,
		BranchID:      tctx.BranchID,
		Type:          input.Type,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedAction: "CREATE",
		UpdatedAction: "CREATE",
	}
	m.CreatedBy = &tctx.UserID
	if err := s.Repo.CreateThread(ctx, m); err != nil {
		return nil, err
	}

	// Add creator as participant
	participant := &db.BunMessageParticipant{
		ID:            uuid.NewString(),
		TenantID:      tctx.TenantID,
		ThreadID:      m.ID,
		ParticipantID: tctx.UserID,
		Role:          "ADMIN",
		LastReadAt:    now,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedAction: "CREATE",
		UpdatedAction: "CREATE",
		CreatedBy:     &tctx.UserID,
	}
	if err := s.Repo.CreateParticipant(ctx, participant); err != nil {
		return nil, err
	}

	// Add additional participants
	for _, p := range input.Participants {
		participant := &db.BunMessageParticipant{
			ID:            uuid.NewString(),
			TenantID:      tctx.TenantID,
			ThreadID:      m.ID,
			ParticipantID: p.ParticipantID,
			Role:          p.Role,
			LastReadAt:    now,
			IsActive:      true,
			CreatedAt:     now,
			UpdatedAt:     now,
			CreatedAction: "CREATE",
			UpdatedAction: "CREATE",
		}
		if tctx.UserID != "" {
			participant.CreatedBy = &tctx.UserID
		}
		if err := s.Repo.CreateParticipant(ctx, participant); err != nil {
			return nil, err
		}
	}

	s.cacheDel(ctx, cache.Key("messaging", "thread", m.ID))
	return m, nil
}

func (s *MessagingService) UpdateThread(ctx context.Context, id string, input model.ThreadUpdateInput) (*db.BunMessageThread, error) {
	existing, err := s.Repo.FindThreadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	if input.Type != nil {
		existing.Type = *input.Type
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}
	existing.UpdatedAt = time.Now()
	existing.UpdatedAction = "UPDATE"

	if err := s.Repo.UpdateThread(ctx, existing); err != nil {
		return nil, err
	}
	s.cacheDel(ctx, cache.Key("messaging", "thread", id))
	return existing, nil
}

func (s *MessagingService) DeleteThread(ctx context.Context, id string) (bool, error) {
	existing, err := s.Repo.FindThreadByID(ctx, id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	existing.IsActive = false
	existing.UpdatedAt = time.Now()
	existing.UpdatedAction = "DELETE"
	if err := s.Repo.UpdateThread(ctx, existing); err != nil {
		return false, err
	}
	s.cacheDel(ctx, cache.Key("messaging", "thread", id))
	return true, nil
}

// Message

func (s *MessagingService) GetMessageByID(ctx context.Context, id string) (*db.BunMessage, error) {
	ck := cache.Key("messaging", "message", id)
	var cached db.BunMessage
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.Repo.FindMessageByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	plaintext, err := s.Encryptor.Decrypt(m.Body, m.Nonce)
	if err != nil {
		return nil, err
	}
	m.Body = string(plaintext)
	m.Nonce = ""
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *MessagingService) GetMessagesByThread(ctx context.Context, threadID string) ([]*db.BunMessage, error) {
	msgs, err := s.Repo.FindMessagesByThread(ctx, threadID)
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		plaintext, err := s.Encryptor.Decrypt(msg.Body, msg.Nonce)
		if err != nil {
			return nil, err
		}
		msg.Body = string(plaintext)
		msg.Nonce = ""
	}
	return msgs, nil
}

func (s *MessagingService) SendMessage(ctx context.Context, input model.MessageInput) (*db.BunMessage, error) {
	tctx := sharedCtx.FromContext(ctx)
	if tctx.UserID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}
	if tctx.TenantID == "" {
		return nil, fmt.Errorf("tenant not identified")
	}

	now := time.Now()

	encryptedBody, nonce, err := s.Encryptor.Encrypt([]byte(input.Body))
	if err != nil {
		return nil, err
	}

	m := &db.BunMessage{
		ID:            uuid.NewString(),
		TenantID:      tctx.TenantID,
		ThreadID:      input.ThreadID,
		SenderID:      tctx.UserID,
		Body:          encryptedBody,
		Nonce:         nonce,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedAction: "CREATE",
		UpdatedAction: "CREATE",
		CreatedBy:     &tctx.UserID,
	}
	if err := s.Repo.CreateMessage(ctx, m); err != nil {
		return nil, err
	}
	s.cacheDel(ctx, cache.Key("messaging", "message", m.ID), cache.Key("messaging", "thread", input.ThreadID, "messages"))

	s.publishMessage(ctx, m, input.Body)
	return m, nil
}

func (s *MessagingService) publishMessage(ctx context.Context, msg *db.BunMessage, plainBody string) {
	if s.Cache == nil {
		return
	}
	participantIDs, err := s.Repo.FindParticipantIDsByThread(ctx, msg.ThreadID)
	if err != nil {
		return
	}
	participantsJSON, _ := json.Marshal(participantIDs)
	s.Cache.XAdd(ctx, &redis.XAddArgs{
		Stream: "messaging:messages",
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{
			"id":           msg.ID,
			"thread_id":    msg.ThreadID,
			"sender_id":    msg.SenderID,
			"body":         plainBody,
			"created_at":   msg.CreatedAt.Format(time.RFC3339),
			"participants": string(participantsJSON),
		},
	})
}

func (s *MessagingService) DecryptMessage(ctx context.Context, msg *db.BunMessage) (string, error) {
	plaintext, err := s.Encryptor.Decrypt(msg.Body, msg.Nonce)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (s *MessagingService) DeleteMessage(ctx context.Context, id string) (bool, error) {
	existing, err := s.Repo.FindMessageByID(ctx, id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	existing.IsActive = false
	existing.UpdatedAt = time.Now()
	existing.UpdatedAction = "DELETE"
	if err := s.Repo.UpdateMessage(ctx, existing); err != nil {
		return false, err
	}
	s.cacheDel(ctx, cache.Key("messaging", "message", id))
	return true, nil
}

// Participant

func (s *MessagingService) GetParticipantsByThread(ctx context.Context, threadID string) ([]*db.BunMessageParticipant, error) {
	return s.Repo.FindParticipantsByThread(ctx, threadID)
}

func (s *MessagingService) AddParticipant(ctx context.Context, input model.ParticipantInput) (*db.BunMessageParticipant, error) {
	tctx := sharedCtx.FromContext(ctx)
	now := time.Now()

	// Check if already a participant
	existing, err := s.Repo.FindParticipantByThreadAndUser(ctx, input.ThreadID, input.ParticipantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if !existing.IsActive {
			existing.IsActive = true
			existing.UpdatedAt = now
			existing.UpdatedAction = "REACTIVATE"
			if err := s.Repo.UpdateParticipant(ctx, existing); err != nil {
				return nil, err
			}
		}
		return existing, nil
	}

	m := &db.BunMessageParticipant{
		ID:            uuid.NewString(),
		TenantID:      tctx.TenantID,
		ThreadID:      input.ThreadID,
		ParticipantID: input.ParticipantID,
		Role:          input.Role,
		LastReadAt:    now,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedAction: "CREATE",
		UpdatedAction: "CREATE",
	}
	if tctx.UserID != "" {
		m.CreatedBy = &tctx.UserID
	}
	if err := s.Repo.CreateParticipant(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MessagingService) RemoveParticipant(ctx context.Context, id string) (bool, error) {
	existing, err := s.Repo.FindParticipantByID(ctx, id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	existing.IsActive = false
	existing.UpdatedAt = time.Now()
	existing.UpdatedAction = "DELETE"
	if err := s.Repo.UpdateParticipant(ctx, existing); err != nil {
		return false, err
	}
	return true, nil
}
