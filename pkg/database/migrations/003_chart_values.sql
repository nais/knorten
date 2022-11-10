-- +goose Up
CREATE TYPE CHART_TYPE AS ENUM ('namespace', 'jupyterhub', 'airflow');

CREATE TABLE chart_global_values (
    "id"         uuid                  DEFAULT uuid_generate_v4(),
    "created"    TIMESTAMPTZ           DEFAULT NOW(),
    "key"        TEXT        NOT NULL,
    "value"      TEXT        NOT NULL,
    "chart_type" CHART_TYPE  NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE chart_team_values (
    "id"         uuid                  DEFAULT uuid_generate_v4(),
    "created"    TIMESTAMPTZ           DEFAULT NOW(),
    "key"        TEXT        NOT NULL,
    "value"      TEXT        NOT NULL,    
    "chart_type" CHART_TYPE  NOT NULL,
    "team"       TEXT        NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_chart_team_values_team
        FOREIGN KEY (team)
            REFERENCES teams (team) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chart_team_values;
DROP TABLE chart_global_values;
DROP TYPE CHART_TYPE;