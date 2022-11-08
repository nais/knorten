package helm

import (
	"context"

	"github.com/nais/knorten/pkg/database"
	"github.com/nais/knorten/pkg/database/gensql"
	"github.com/nais/knorten/pkg/helm"
	"helm.sh/helm/v3/pkg/chart"
)

type Jupyterhub struct {
	team string
	repo *database.Repo
}

func NewJupyterhub(team string, repo *database.Repo) *Jupyterhub {
	return &Jupyterhub{
		team: team,
		repo: repo,
	}
}

func (j *Jupyterhub) Chart(ctx context.Context) (*chart.Chart, error) {
	chart, err := helm.FetchChart("jupyterhub", "jupyterhub", "0.11.1")
	if err != nil {
		return nil, err
	}

	err = j.mergeValues(ctx, chart.Values)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func (j *Jupyterhub) mergeValues(ctx context.Context, defaultValues map[string]any) error {
	values, err := j.globalValues(ctx)
	if err != nil {
		return err
	}

	values, err = j.enrichWithTeamValues(ctx, values)
	if err != nil {
		return err
	}

	for key, value := range values {
		keyPath := helm.KeySplitHandleEscape(key)
		helm.SetChartValue(keyPath, value, defaultValues)
	}

	return nil
}

func (j *Jupyterhub) globalValues(ctx context.Context) (map[string]any, error) {
	dbValues, err := j.repo.GlobalValuesGet(ctx, gensql.ChartTypeJupyterhub)
	if err != nil {
		return map[string]any{}, err
	}

	values := map[string]any{}
	for _, v := range dbValues {
		values[v.Key], err = helm.ParseValue(v.Value)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (j *Jupyterhub) enrichWithTeamValues(ctx context.Context, values map[string]any) (map[string]any, error) {
	dbValues, err := j.repo.TeamValuesGet(ctx, gensql.ChartTypeJupyterhub, j.team)
	if err != nil {
		return map[string]any{}, err
	}

	for _, v := range dbValues {
		values[v.Key], err = helm.ParseValue(v.Value)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}
