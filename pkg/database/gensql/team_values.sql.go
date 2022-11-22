// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: team_values.sql

package gensql

import (
	"context"
)

const appDelete = `-- name: AppDelete :exec
DELETE FROM chart_team_values
WHERE team_id = $1 AND chart_type = $2
`

type AppDeleteParams struct {
	TeamID    string
	ChartType ChartType
}

func (q *Queries) AppDelete(ctx context.Context, arg AppDeleteParams) error {
	_, err := q.db.ExecContext(ctx, appDelete, arg.TeamID, arg.ChartType)
	return err
}

const appsForTeamGet = `-- name: AppsForTeamGet :many
SELECT DISTINCT ON (chart_type) chart_type
FROM chart_team_values
WHERE team_id = $1
`

func (q *Queries) AppsForTeamGet(ctx context.Context, teamID string) ([]ChartType, error) {
	rows, err := q.db.QueryContext(ctx, appsForTeamGet, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartType{}
	for rows.Next() {
		var chart_type ChartType
		if err := rows.Scan(&chart_type); err != nil {
			return nil, err
		}
		items = append(items, chart_type)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const teamValueInsert = `-- name: TeamValueInsert :exec
INSERT INTO chart_team_values ("key",
                               "value",
                               "team_id",
                               "chart_type")
VALUES ($1,
        $2,
        $3,
        $4)
`

type TeamValueInsertParams struct {
	Key       string
	Value     string
	TeamID    string
	ChartType ChartType
}

func (q *Queries) TeamValueInsert(ctx context.Context, arg TeamValueInsertParams) error {
	_, err := q.db.ExecContext(ctx, teamValueInsert,
		arg.Key,
		arg.Value,
		arg.TeamID,
		arg.ChartType,
	)
	return err
}

const teamValuesGet = `-- name: TeamValuesGet :many
SELECT DISTINCT ON ("key") id, created, key, value, chart_type, team_id
FROM chart_team_values
WHERE chart_type = $1
  AND team_id = $2
ORDER BY "key", "created" DESC
`

type TeamValuesGetParams struct {
	ChartType ChartType
	TeamID    string
}

func (q *Queries) TeamValuesGet(ctx context.Context, arg TeamValuesGetParams) ([]ChartTeamValue, error) {
	rows, err := q.db.QueryContext(ctx, teamValuesGet, arg.ChartType, arg.TeamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartTeamValue{}
	for rows.Next() {
		var i ChartTeamValue
		if err := rows.Scan(
			&i.ID,
			&i.Created,
			&i.Key,
			&i.Value,
			&i.ChartType,
			&i.TeamID,
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
