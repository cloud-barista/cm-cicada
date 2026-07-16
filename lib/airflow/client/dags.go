package client

import (
	"context"
	"net/url"
)

// GetDAG fetches one DAG. A missing DAG returns an error for which
// IsNotFound reports true.
func (c *Client) GetDAG(ctx context.Context, dagID string) (DAG, error) {
	var out DAG
	err := c.do(ctx, "GET", "/api/v2/dags/"+pathEscape(dagID), nil, nil, &out)
	return out, err
}

// DAGExists reports whether the DAG is registered with Airflow.
func (c *Client) DAGExists(ctx context.Context, dagID string) (bool, error) {
	_, err := c.GetDAG(ctx, dagID)
	if err == nil {
		return true, nil
	}
	if IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// DeleteDAG removes a DAG and its history from Airflow.
func (c *Client) DeleteDAG(ctx context.Context, dagID string) error {
	return c.do(ctx, "DELETE", "/api/v2/dags/"+pathEscape(dagID), nil, nil, nil)
}

// ClearTaskInstances re-queues task instances of a DAG run.
func (c *Client) ClearTaskInstances(ctx context.Context, dagID string, body ClearTaskInstancesRequest) (TaskInstanceCollection, error) {
	var out TaskInstanceCollection
	err := c.do(ctx, "POST", "/api/v2/dags/"+pathEscape(dagID)+"/clearTaskInstances", nil, body, &out)
	return out, err
}

// GetImportErrors lists DAG import errors reported by the DAG processor.
func (c *Client) GetImportErrors(ctx context.Context) (ImportErrorCollection, error) {
	var out ImportErrorCollection
	err := c.do(ctx, "GET", "/api/v2/importErrors", nil, nil, &out)
	return out, err
}

// GetEventLogs returns the event log, newest first.
//
// dagRunID and taskID are optional; empty values are omitted from the filter.
func (c *Client) GetEventLogs(ctx context.Context, dagID, dagRunID, taskID string, limit, offset int) (EventLogCollection, error) {
	q := url.Values{}
	q.Set("limit", itoa(limit))
	q.Set("offset", itoa(offset))
	q.Set("order_by", "-when")
	if dagID != "" {
		q.Set("dag_id", dagID)
	}
	if dagRunID != "" {
		q.Set("run_id", dagRunID)
	}
	if taskID != "" {
		q.Set("task_id", taskID)
	}
	// UI-only bookkeeping events; excluded so callers see real activity.
	q.Set("excluded_events", "gantt,landing_times,tries,duration,calendar,graph,grid,tree,tree_data")

	var out EventLogCollection
	err := c.do(ctx, "GET", "/api/v2/eventLogs", q, nil, &out)
	return out, err
}
