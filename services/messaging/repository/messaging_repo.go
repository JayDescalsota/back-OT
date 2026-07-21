package repository

import (
	"context"
	"database/sql"

	"github.com/clinicmanager/services/messaging/db"
	shareddb "github.com/clinicmanager/shared/db"
)

type MessagingRepository struct {
	db       *shareddb.ScopedDB
	tenantdb *shareddb.ScopedDB
	alldb    *shareddb.ScopedDB
}

func NewMessagingRepository(dbs *shareddb.DBSet) *MessagingRepository {
	return &MessagingRepository{
		db:       dbs.DB,
		tenantdb: dbs.TenantDB,
		alldb:    dbs.AllDB,
	}
}

// Thread

func (r *MessagingRepository) FindThreadByID(ctx context.Context, id string) (*db.BunMessageThread, error) {
	var m db.BunMessageThread
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MessagingRepository) FindThreadsByParticipant(ctx context.Context, participantID string) ([]*db.BunMessageThread, error) {
	var threads []*db.BunMessageThread
	err := r.db.NewSelect(ctx, &threads).
		Join("JOIN message_participant mp ON mp.thread_id = message_thread.id").
		Where("mp.participant_id = ? AND mp.is_active = true", participantID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r *MessagingRepository) ListThreadsByBranch(ctx context.Context, branchID string) ([]*db.BunMessageThread, error) {
	var threads []*db.BunMessageThread
	err := r.tenantdb.NewSelect(ctx, &threads).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r *MessagingRepository) CreateThread(ctx context.Context, m *db.BunMessageThread) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *MessagingRepository) UpdateThread(ctx context.Context, m *db.BunMessageThread) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// Message

func (r *MessagingRepository) FindMessageByID(ctx context.Context, id string) (*db.BunMessage, error) {
	var m db.BunMessage
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MessagingRepository) FindMessagesByThread(ctx context.Context, threadID string) ([]*db.BunMessage, error) {
	var msgs []*db.BunMessage
	err := r.db.NewSelect(ctx, &msgs).Where("thread_id = ?", threadID).Order("created_at ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *MessagingRepository) CreateMessage(ctx context.Context, m *db.BunMessage) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *MessagingRepository) UpdateMessage(ctx context.Context, m *db.BunMessage) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// Participant

func (r *MessagingRepository) FindParticipantByID(ctx context.Context, id string) (*db.BunMessageParticipant, error) {
	var m db.BunMessageParticipant
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MessagingRepository) FindParticipantsByThread(ctx context.Context, threadID string) ([]*db.BunMessageParticipant, error) {
	var participants []*db.BunMessageParticipant
	err := r.db.NewSelect(ctx, &participants).Where("thread_id = ? AND is_active = true", threadID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (r *MessagingRepository) FindParticipantByThreadAndUser(ctx context.Context, threadID, participantID string) (*db.BunMessageParticipant, error) {
	var m db.BunMessageParticipant
	err := r.db.NewSelect(ctx, &m).Where("thread_id = ? AND participant_id = ?", threadID, participantID).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MessagingRepository) CreateParticipant(ctx context.Context, m *db.BunMessageParticipant) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *MessagingRepository) UpdateParticipant(ctx context.Context, m *db.BunMessageParticipant) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}
