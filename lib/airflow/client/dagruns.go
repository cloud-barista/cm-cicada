package client

import (
	"context"
	"net/url"
	"strconv"
)

func itoa(i int) string { return strconv.Itoa(i) }

// TriggerDAGRun starts a DAG run.
//
// conf may be nil. The request always carries logical_date (null for a manual
// run): Airflow 3 rejects a body without the key with 422.
func (c *Client) TriggerDAGRun(ctx context.Context, dagID string, conf map[string]interface{}) (DAGRun, error) {
	body := TriggerDAGRunRequest{
		LogicalDate: nil,
		Conf:        conf,
	}

	var out DAGRun
	err := c.do(ctx, "POST", "/api/v2/dags/"+pathEscape(dagID)+"/dagRuns", nil, body, &out)
	return out, err
}

// GetDAGRuns lists runs of a DAG. states filters by run state; nil means all.
func (c *Client) GetDAGRuns(ctx context.Context, dagID string, states []string) (DAGRunCollection, error) {
	q := url.Values{}
	for _, s := range states {
		q.Add("state", s)
	}

	var out DAGRunCollection
	err := c.do(ctx, "GET", "/api/v2/dags/"+pathEscape(dagID)+"/dagRuns", q, nil, &out)
	return out, err
}

// GetDAGRun fetches a single DAG run.
func (c *Client) GetDAGRun(ctx context.Context, dagID, dagRunID string) (DAGRun, error) {
	var out DAGRun
	err := c.do(ctx, "GET", dagRunPath(dagID, dagRunID), nil, nil, &out)
	return out, err
}

func dagRunPath(dagID, dagRunID string) string {
	return "/api/v2/dags/" + pathEscape(dagID) + "/dagRuns/" + pathEscape(dagRunID)
}
