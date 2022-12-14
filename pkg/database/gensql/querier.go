// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package gensql

import (
	"context"
)

type Querier interface {
	AppDelete(ctx context.Context, arg AppDeleteParams) error
	AppsForTeamGet(ctx context.Context, teamID string) ([]ChartType, error)
	ClearPendingUpgradeLocks(ctx context.Context) error
	GlobalValueDelete(ctx context.Context, arg GlobalValueDeleteParams) error
	GlobalValueInsert(ctx context.Context, arg GlobalValueInsertParams) error
	GlobalValuesGet(ctx context.Context, chartType ChartType) ([]ChartGlobalValue, error)
	SessionCreate(ctx context.Context, arg SessionCreateParams) error
	SessionDelete(ctx context.Context, token string) error
	SessionGet(ctx context.Context, token string) (Session, error)
	TeamCreate(ctx context.Context, arg TeamCreateParams) error
	TeamDelete(ctx context.Context, id string) error
	TeamGet(ctx context.Context, slug string) (TeamGetRow, error)
	TeamSetPendingAirflowUpgrade(ctx context.Context, arg TeamSetPendingAirflowUpgradeParams) error
	TeamSetPendingJupyterUpgrade(ctx context.Context, arg TeamSetPendingJupyterUpgradeParams) error
	TeamUpdate(ctx context.Context, arg TeamUpdateParams) error
	TeamValueInsert(ctx context.Context, arg TeamValueInsertParams) error
	TeamValuesGet(ctx context.Context, arg TeamValuesGetParams) ([]ChartTeamValue, error)
	TeamsForAppGet(ctx context.Context, chartType ChartType) ([]string, error)
	TeamsForUserGet(ctx context.Context, email string) ([]TeamsForUserGetRow, error)
	TeamsGet(ctx context.Context) ([]Team, error)
}

var _ Querier = (*Queries)(nil)
