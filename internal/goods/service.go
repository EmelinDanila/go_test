package goods

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EmelinDanila/go_test/internal/cache"
	"github.com/EmelinDanila/go_test/internal/db"
	natslog "github.com/EmelinDanila/go_test/internal/nats"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = fmt.Errorf("errors.common.notFound")

func GetWithCache(ctx context.Context, id int) (*Good, error) {
	key := fmt.Sprintf("good:%d", id)
	var g Good

	err := cache.GetJSON(ctx, key, &g)
	if err == nil {
		return &g, nil
	}

	gPtr, err := Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cache.SetJSON(ctx, key, gPtr, time.Minute)
	return gPtr, nil
}

func getMaxPriority(ctx context.Context, tx pgx.Tx) (int, error) {
	var max int
	err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(priority), 0) FROM goods`).Scan(&max)
	return max, err
}

func Create(ctx context.Context, projectID int, name, description string) (*Good, error) {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	priority, err := getMaxPriority(ctx, tx)
	if err != nil {
		return nil, err
	}
	priority += 1

	var g Good
	err = tx.QueryRow(ctx, `
		INSERT INTO goods(project_id, name, description, priority, removed)
		VALUES ($1, $2, $3, $4, FALSE)
		RETURNING id, project_id, name, description, priority, removed, created_at
	`, projectID, name, description, priority).Scan(
		&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = natslog.Publish(ctx, natslog.Event{
		ID:          uint32(g.ID),
		ProjectID:   uint32(g.ProjectID),
		Name:        g.Name,
		Description: g.Description,
		Priority:    g.Priority,
		Removed:     g.Removed,
		EventTime:   time.Now(),
	})

	return &g, nil
}

func Get(ctx context.Context, id int) (*Good, error) {
	var g Good
	err := db.DB.QueryRow(ctx, `
		SELECT id, project_id, name, description, priority, removed, created_at
		FROM goods
		WHERE id = $1
	`, id).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func Update(ctx context.Context, id int, name, description string) (*Good, error) {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM goods WHERE id = $1 FOR UPDATE)`, id).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}

	var g Good
	err = tx.QueryRow(ctx, `
		UPDATE goods SET name = $1, description = $2
		WHERE id = $3
		RETURNING id, project_id, name, description, priority, removed, created_at
	`, name, description, id).Scan(
		&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	cache.Delete(ctx, fmt.Sprintf("good:%d", id))

	_ = natslog.Publish(ctx, natslog.Event{
		ID:          uint32(g.ID),
		ProjectID:   uint32(g.ProjectID),
		Name:        g.Name,
		Description: g.Description,
		Priority:    g.Priority,
		Removed:     g.Removed,
		EventTime:   time.Now(),
	})

	return &g, tx.Commit(ctx)
}

func Delete(ctx context.Context, id int) error {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var g Good
	err = tx.QueryRow(ctx, `
		SELECT id, project_id, name, description, priority, removed, created_at
		FROM goods WHERE id = $1 FOR UPDATE
	`, id).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM goods WHERE id = $1`, id)
	if err != nil {
		return err
	}

	cache.Delete(ctx, fmt.Sprintf("good:%d", id))

	_ = natslog.Publish(ctx, natslog.Event{
		ID:          uint32(g.ID),
		ProjectID:   uint32(g.ProjectID),
		Name:        g.Name,
		Description: g.Description,
		Priority:    g.Priority,
		Removed:     true,
		EventTime:   time.Now(),
	})

	return tx.Commit(ctx)
}
