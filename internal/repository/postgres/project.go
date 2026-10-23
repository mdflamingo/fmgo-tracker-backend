package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mdflamingo/fgo-tracker-backend/internal/model"
)

func (d *DBStorage) CreateProject(project model.ProjectCreate, creator model.Identity, userProjectList []model.ProjectUserCreate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := d.EnsureUserTx(ctx, tx, creator); err != nil {
		return err
	}

	userIDs := make([]uuid.UUID, 0, len(userProjectList))
	seen := make(map[uuid.UUID]struct{}, len(userProjectList))
	for _, up := range userProjectList {
		if up.UserId == uuid.Nil {
			continue
		}
		if _, ok := seen[up.UserId]; ok {
			continue
		}
		seen[up.UserId] = struct{}{}
		userIDs = append(userIDs, up.UserId)
	}

	missing, err := d.MissingUserIdsTx(ctx, tx, userIDs)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return &model.UnknownUsersError{IDs: missing}
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO project (id, name)
         VALUES ($1, $2)
		 RETURNING id`,
		project.Id, project.Name)

	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	if len(userProjectList) > 0 {
		ids := make([]uuid.UUID, len(userProjectList))
		projectIDs := make([]uuid.UUID, len(userProjectList))
		userIDs := make([]uuid.UUID, len(userProjectList))
		roles := make([]string, len(userProjectList))

		for i, ut := range userProjectList {
			ids[i] = ut.Id
			projectIDs[i] = project.Id
			userIDs[i] = ut.UserId
			roles[i] = string(ut.Role)
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO user_project (id, project_id, user_id, role)
			 SELECT 
				unnest($1::uuid[]),
				unnest($2::uuid[]),
				unnest($3::uuid[]),
				unnest($4::project_role[])`,
			ids, projectIDs, userIDs, roles,
		)
		if err != nil {
			return fmt.Errorf("failed to save user tasks: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *DBStorage) GetProjectList() ([]model.ProjectListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := d.pool.Query(ctx,
		`SELECT id, name
			FROM project
			ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	projectsList := make([]model.ProjectListResponse, 0)

	for rows.Next() {
		var project model.ProjectListResponse
		if err := rows.Scan(&project.Id, &project.Name); err != nil {
			return nil, fmt.Errorf("data scan error: %w", err)
		}
		projectsList = append(projectsList, project)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows processing error: %w", err)
	}

	return projectsList, nil
}
