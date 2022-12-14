// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: teams.sql

package gensql

import (
	"context"

	"github.com/lib/pq"
)

const clearPendingUpgradeLocks = `-- name: ClearPendingUpgradeLocks :exec
UPDATE teams
SET pending_jupyter_upgrade = false, pending_airflow_upgrade = false
`

func (q *Queries) ClearPendingUpgradeLocks(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, clearPendingUpgradeLocks)
	return err
}

const teamCreate = `-- name: TeamCreate :exec
INSERT INTO teams ("id", "users", "slug")
VALUES ($1, $2, $3)
`

type TeamCreateParams struct {
	ID    string
	Users []string
	Slug  string
}

func (q *Queries) TeamCreate(ctx context.Context, arg TeamCreateParams) error {
	_, err := q.db.ExecContext(ctx, teamCreate, arg.ID, pq.Array(arg.Users), arg.Slug)
	return err
}

const teamDelete = `-- name: TeamDelete :exec
DELETE
FROM teams
WHERE id = $1
`

func (q *Queries) TeamDelete(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, teamDelete, id)
	return err
}

const teamGet = `-- name: TeamGet :one
SELECT id, users, slug, pending_jupyter_upgrade, pending_airflow_upgrade
FROM teams
WHERE slug = $1
`

type TeamGetRow struct {
	ID                    string
	Users                 []string
	Slug                  string
	PendingJupyterUpgrade bool
	PendingAirflowUpgrade bool
}

func (q *Queries) TeamGet(ctx context.Context, slug string) (TeamGetRow, error) {
	row := q.db.QueryRowContext(ctx, teamGet, slug)
	var i TeamGetRow
	err := row.Scan(
		&i.ID,
		pq.Array(&i.Users),
		&i.Slug,
		&i.PendingJupyterUpgrade,
		&i.PendingAirflowUpgrade,
	)
	return i, err
}

const teamSetPendingAirflowUpgrade = `-- name: TeamSetPendingAirflowUpgrade :exec
UPDATE teams
SET pending_airflow_upgrade = $1
WHERE id = $2
`

type TeamSetPendingAirflowUpgradeParams struct {
	PendingAirflowUpgrade bool
	ID                    string
}

func (q *Queries) TeamSetPendingAirflowUpgrade(ctx context.Context, arg TeamSetPendingAirflowUpgradeParams) error {
	_, err := q.db.ExecContext(ctx, teamSetPendingAirflowUpgrade, arg.PendingAirflowUpgrade, arg.ID)
	return err
}

const teamSetPendingJupyterUpgrade = `-- name: TeamSetPendingJupyterUpgrade :exec
UPDATE teams
SET pending_jupyter_upgrade = $1
WHERE id = $2
`

type TeamSetPendingJupyterUpgradeParams struct {
	PendingJupyterUpgrade bool
	ID                    string
}

func (q *Queries) TeamSetPendingJupyterUpgrade(ctx context.Context, arg TeamSetPendingJupyterUpgradeParams) error {
	_, err := q.db.ExecContext(ctx, teamSetPendingJupyterUpgrade, arg.PendingJupyterUpgrade, arg.ID)
	return err
}

const teamUpdate = `-- name: TeamUpdate :exec
UPDATE teams
SET users = $1
WHERE id = $2
`

type TeamUpdateParams struct {
	Users []string
	ID    string
}

func (q *Queries) TeamUpdate(ctx context.Context, arg TeamUpdateParams) error {
	_, err := q.db.ExecContext(ctx, teamUpdate, pq.Array(arg.Users), arg.ID)
	return err
}

const teamsForUserGet = `-- name: TeamsForUserGet :many
SELECT id, slug
FROM teams
WHERE $1::TEXT = ANY ("users")
`

type TeamsForUserGetRow struct {
	ID   string
	Slug string
}

func (q *Queries) TeamsForUserGet(ctx context.Context, email string) ([]TeamsForUserGetRow, error) {
	rows, err := q.db.QueryContext(ctx, teamsForUserGet, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TeamsForUserGetRow{}
	for rows.Next() {
		var i TeamsForUserGetRow
		if err := rows.Scan(&i.ID, &i.Slug); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const teamsGet = `-- name: TeamsGet :many
select id, slug, users, created, pending_jupyter_upgrade, pending_airflow_upgrade
from teams
ORDER BY slug
`

func (q *Queries) TeamsGet(ctx context.Context) ([]Team, error) {
	rows, err := q.db.QueryContext(ctx, teamsGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Team{}
	for rows.Next() {
		var i Team
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			pq.Array(&i.Users),
			&i.Created,
			&i.PendingJupyterUpgrade,
			&i.PendingAirflowUpgrade,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
