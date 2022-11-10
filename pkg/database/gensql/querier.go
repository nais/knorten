// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package gensql

import (
	"context"
)

type Querier interface {
	AppsForTeamGet(ctx context.Context, team string) ([]AppsForTeamGetRow, error)
	GlobalValueInsert(ctx context.Context, arg GlobalValueInsertParams) error
	GlobalValuesGet(ctx context.Context, chartType ChartType) ([]ChartGlobalValue, error)
	SessionCreate(ctx context.Context, arg SessionCreateParams) error
	SessionDelete(ctx context.Context, token string) error
	SessionGet(ctx context.Context, token string) (Session, error)
	TeamCreate(ctx context.Context, arg TeamCreateParams) error
	TeamDelete(ctx context.Context, team string) error
	TeamGet(ctx context.Context, team string) (TeamGetRow, error)
	TeamUpdate(ctx context.Context, arg TeamUpdateParams) error
	TeamValueInsert(ctx context.Context, arg TeamValueInsertParams) error
	TeamValuesGet(ctx context.Context, arg TeamValuesGetParams) ([]ChartTeamValue, error)
	TeamsForUserGet(ctx context.Context, email string) ([]string, error)
	TeamsGet(ctx context.Context) ([]TeamsGetRow, error)
}

var _ Querier = (*Queries)(nil)
