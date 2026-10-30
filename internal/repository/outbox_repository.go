package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"spot-solution/internal/database"
	"spot-solution/internal/models"
)

type OutboxRepository struct {
	db *database.DB
}

func NewOutboxRepository(db *database.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) QueueEmailTx(ctx context.Context, tx *sql.Tx, item *models.EmailOutboxItem) error {
	query := `
		INSERT INTO email_outbox (
			id, submission_id, recipient_email, recipient_name, subject, body_html, status, retry_count, max_retries, next_retry_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := tx.ExecContext(ctx, query,
		item.ID, item.SubmissionID, item.RecipientEmail, item.RecipientName,
		item.Subject, item.BodyHTML, item.Status, item.RetryCount,
		item.MaxRetries, item.NextRetryAt, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to queue email into outbox: %w", err)
	}
	return nil
}

func (r *OutboxRepository) FetchPendingBatch(ctx context.Context, limit int) ([]models.EmailOutboxItem, error) {
	query := `
		SELECT id, submission_id, recipient_email, recipient_name, subject, body_html, status, retry_count, max_retries, next_retry_at, last_error, created_at, sent_at
		FROM email_outbox
		WHERE status IN ('pending', 'failed') AND next_retry_at <= $1 AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox emails: %w", err)
	}
	defer rows.Close()

	var items []models.EmailOutboxItem
	for rows.Next() {
		var item models.EmailOutboxItem
		var statusStr string
		err := rows.Scan(
			&item.ID, &item.SubmissionID, &item.RecipientEmail, &item.RecipientName,
			&item.Subject, &item.BodyHTML, &statusStr, &item.RetryCount,
			&item.MaxRetries, &item.NextRetryAt, &item.LastError,
			&item.CreatedAt, &item.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning outbox item: %w", err)
		}
		item.Status = models.OutboxStatus(statusStr)
		items = append(items, item)
	}

	return items, nil
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE email_outbox SET status = 'sent', sent_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextRetry time.Time) error {
	query := `
		UPDATE email_outbox
		SET status = 'failed', retry_count = retry_count + 1, last_error = $1, next_retry_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, errMsg, nextRetry, id)
	return err
}
