// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package gensql

import (
	"context"
)

type Querier interface {
	AppDelete(ctx context.Context, arg AppDeleteParams) error
	AppsForTeamGet(ctx context.Context, teamID string) ([]ChartType, error)
	GlobalValueInsert(ctx context.Context, arg GlobalValueInsertParams) error
	GlobalValuesGet(ctx context.Context, chartType ChartType) ([]ChartGlobalValue, error)
	SessionCreate(ctx context.Context, arg SessionCreateParams) error
	SessionDelete(ctx context.Context, token string) error
	SessionGet(ctx context.Context, token string) (Session, error)
	TeamCreate(ctx context.Context, arg TeamCreateParams) error
	TeamDelete(ctx context.Context, id string) error
	TeamGet(ctx context.Context, slug string) (TeamGetRow, error)
	TeamUpdate(ctx context.Context, arg TeamUpdateParams) error
	TeamValueInsert(ctx context.Context, arg TeamValueInsertParams) error
	TeamValuesGet(ctx context.Context, arg TeamValuesGetParams) ([]ChartTeamValue, error)
	TeamsForUserGet(ctx context.Context, email string) ([]TeamsForUserGetRow, error)
}

var _ Querier = (*Queries)(nil)
