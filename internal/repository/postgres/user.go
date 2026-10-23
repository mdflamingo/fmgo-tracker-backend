package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mdflamingo/fgo-tracker-backend/internal/model"
)

func (d *DBStorage) GetUserList() ([]model.UserDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := d.pool.Query(ctx,
		`SELECT id, username, email
			FROM "user"
			ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	usersList := make([]model.UserDB, 0)

	for rows.Next() {
		var (
			user  model.UserDB
			email *string
		)
		if err := rows.Scan(&user.Id, &user.Username, &email); err != nil {
			return nil, fmt.Errorf("data scan error: %w", err)
		}
		user.Email = derefStr(email)
		usersList = append(usersList, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows processing error: %w", err)
	}

	return usersList, nil
}

// EnsureUserTx создаёт строку в "user" для владельца токена, если её ещё нет.
// Вызывается внутри той же транзакции, что и вставка связей user_task/user_project,
// иначе FK на user_id упадёт.
func (d *DBStorage) EnsureUserTx(ctx context.Context, tx pgx.Tx, identity model.Identity) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO "user" (id, username, email)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO NOTHING`,
		identity.ID, identity.Username, identity.Email)
	if err != nil {
		return fmt.Errorf("failed to provision user: %w", err)
	}

	return nil
}

// MissingUserIdsTx возвращает из ids те, для которых нет строки в "user".
// Вызывается до вставки связей, чтобы отдать 422 вместо 500 по FK.
func (d *DBStorage) MissingUserIdsTx(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := tx.Query(ctx,
		`SELECT u.id
		 FROM unnest($1::uuid[]) AS u(id)
		 WHERE NOT EXISTS (SELECT 1 FROM "user" x WHERE x.id = u.id)`,
		ids)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user ids: %w", err)
	}
	defer rows.Close()

	missing := make([]uuid.UUID, 0, len(ids))
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("data scan error: %w", err)
		}
		missing = append(missing, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows processing error: %w", err)
	}

	return missing, nil
}
