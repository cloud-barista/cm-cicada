package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// GetTaskInstances lists the task instances of a DAG run.
func (c *Client) GetTaskInstances(ctx context.Context, dagID, dagRunID string) (TaskInstanceCollection, error) {
	var out TaskInstanceCollection
	err := c.do(ctx, "GET", dagRunPath(dagID, dagRunID)+"/taskInstances", nil, nil, &out)
	return out, err
}

// GetTaskLog returns a task attempt's log as plain text.
//
// Airflow 3 returns structured records instead of the Airflow 2 API's single
// string, so the entries are rendered back into lines here and callers keep
// dealing in text.
func (c *Client) GetTaskLog(ctx context.Context, dagID, dagRunID, taskID string, tryNumber int) (string, error) {
	q := url.Values{}
	q.Set("full_content", "true")

	path := dagRunPath(dagID, dagRunID) + "/taskInstances/" + pathEscape(taskID) + "/logs/" + itoa(tryNumber)

	raw, err := c.doRaw(ctx, "GET", path, q)
	if err != nil {
		return "", err
	}

	var resp struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", err
	}
	return renderLogContent(resp.Content), nil
}

// renderLogContent flattens the log payload into text. content is either an
// array of StructuredLogMessage objects or an array of plain strings.
func renderLogContent(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}

	var messages []StructuredLogMessage
	if err := json.Unmarshal(content, &messages); err == nil {
		var b strings.Builder
		for _, m := range messages {
			if m.Timestamp != nil {
				b.WriteString("[")
				b.WriteString(m.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"))
				b.WriteString("] ")
			}
			b.WriteString(m.Event)
			b.WriteString("\n")
		}
		return b.String()
	}

	var lines []string
	if err := json.Unmarshal(content, &lines); err == nil {
		return strings.Join(lines, "\n")
	}

	// Unknown shape: hand back the raw JSON rather than silently dropping logs.
	return string(content)
}

// GetXComEntry reads one XCom value of a task instance.
func (c *Client) GetXComEntry(ctx context.Context, dagID, dagRunID, taskID, xcomKey string) (XComEntry, error) {
	var out XComEntry
	path := dagRunPath(dagID, dagRunID) + "/taskInstances/" + pathEscape(taskID) + "/xcomEntries/" + pathEscape(xcomKey)
	err := c.do(ctx, "GET", path, nil, nil, &out)
	return out, err
}
