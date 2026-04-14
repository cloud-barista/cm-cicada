package common

import "github.com/cloud-barista/cm-cicada/pkg/api/rest/model"

// WorkflowDagID returns the Airflow DAG ID for the given workflow.
// workflow_key is used when set (stable runtime key); falls back to workflow.ID for legacy records.
func WorkflowDagID(workflow *model.Workflow) string {
	if workflow.WorkflowKey != "" {
		return workflow.WorkflowKey
	}
	return workflow.ID
}
