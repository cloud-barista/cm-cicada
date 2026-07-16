package client

import (
	"context"
	"os"
	"testing"
	"time"
)

// Exercises the client against a real Airflow 3 server. Skipped unless
// AIRFLOW_LIVE_TEST_ADDR is set, e.g.
//
//	AIRFLOW_LIVE_TEST_ADDR=127.0.0.1:8080 go test ./lib/airflow/client/ -run TestLive -v
func liveClient(t *testing.T) *Client {
	t.Helper()

	addr := os.Getenv("AIRFLOW_LIVE_TEST_ADDR")
	if addr == "" {
		t.Skip("AIRFLOW_LIVE_TEST_ADDR not set; skipping live Airflow test")
	}

	user := os.Getenv("AIRFLOW_LIVE_TEST_USER")
	if user == "" {
		user = "airflow"
	}
	pass := os.Getenv("AIRFLOW_LIVE_TEST_PASSWORD")
	if pass == "" {
		pass = "airflow_pass"
	}

	return New(Config{
		Address:  addr,
		Username: user,
		Password: pass,
		Timeout:  30 * time.Second,
	})
}

func liveCtx(t *testing.T) (context.Context, func()) {
	t.Helper()
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func TestLiveHealthAndAuth(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	if err := c.Health(ctx); err != nil {
		t.Fatalf("Health: %v", err)
	}

	// Proves the JWT handshake against POST /auth/token works; Airflow 3
	// rejects the Basic auth the previous client used.
	token, err := c.accessToken(ctx, false)
	if err != nil {
		t.Fatalf("accessToken: %v", err)
	}
	if token == "" {
		t.Fatal("accessToken returned an empty token")
	}
	if c.tokenExpiry.Before(time.Now()) {
		t.Errorf("token expiry %v is already past; exp claim was not parsed", c.tokenExpiry)
	}
	t.Logf("token acquired, expires %s", c.tokenExpiry.Format(time.RFC3339))
}

func TestLiveDAGs(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	dag, err := c.GetDAG(ctx, "monitor_dag")
	if err != nil {
		t.Fatalf("GetDAG: %v", err)
	}
	if dag.DagID != "monitor_dag" {
		t.Errorf("DagID = %q, want monitor_dag", dag.DagID)
	}

	exists, err := c.DAGExists(ctx, "monitor_dag")
	if err != nil || !exists {
		t.Errorf("DAGExists(monitor_dag) = %v, %v; want true, nil", exists, err)
	}

	// A missing DAG must be a clean false, not an error: RunDAG relies on this
	// to report "provided dag_id is not exist".
	exists, err = c.DAGExists(ctx, "no_such_dag_xyz")
	if err != nil {
		t.Errorf("DAGExists(missing) returned error: %v", err)
	}
	if exists {
		t.Error("DAGExists(no_such_dag_xyz) = true, want false")
	}

	if _, err := c.GetDAG(ctx, "no_such_dag_xyz"); !IsNotFound(err) {
		t.Errorf("IsNotFound on missing DAG = false; err = %v", err)
	}
}

func TestLiveTriggerAndInspectRun(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	conf := map[string]interface{}{"probe": "live-test"}
	run, err := c.TriggerDAGRun(ctx, "templates", conf)
	if err != nil {
		t.Fatalf("TriggerDAGRun: %v", err)
	}
	if run.DagRunID == "" {
		t.Fatal("TriggerDAGRun returned an empty dag_run_id")
	}
	t.Logf("triggered run %s (state=%s)", run.DagRunID, run.State)

	// Round-tripping the run id proves path escaping holds: run ids contain
	// '+' and ':', which a raw URL would mangle.
	got, err := c.GetDAGRun(ctx, "templates", run.DagRunID)
	if err != nil {
		t.Fatalf("GetDAGRun(%q): %v", run.DagRunID, err)
	}
	if got.DagRunID != run.DagRunID {
		t.Errorf("GetDAGRun returned %q, want %q", got.DagRunID, run.DagRunID)
	}
	if got.Conf["probe"] != "live-test" {
		t.Errorf("conf round-trip failed: %#v", got.Conf)
	}

	runs, err := c.GetDAGRuns(ctx, "templates", nil)
	if err != nil {
		t.Fatalf("GetDAGRuns: %v", err)
	}
	if runs.TotalEntries == 0 {
		t.Error("GetDAGRuns returned no runs")
	}

	// State filtering backs GetDagStatus.
	for _, state := range DagRunStates {
		filtered, err := c.GetDAGRuns(ctx, "templates", []string{state})
		if err != nil {
			t.Fatalf("GetDAGRuns(state=%s): %v", state, err)
		}
		for _, r := range filtered.DagRuns {
			if r.State != state {
				t.Errorf("state filter %q returned a run in state %q", state, r.State)
			}
		}
	}
}

func TestLiveTaskInstancesLogsAndXCom(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	runs, err := c.GetDAGRuns(ctx, "monitor_dag", nil)
	if err != nil {
		t.Fatalf("GetDAGRuns: %v", err)
	}
	if len(runs.DagRuns) == 0 {
		t.Skip("no monitor_dag runs to inspect")
	}
	runID := runs.DagRuns[len(runs.DagRuns)-1].DagRunID

	instances, err := c.GetTaskInstances(ctx, "monitor_dag", runID)
	if err != nil {
		t.Fatalf("GetTaskInstances: %v", err)
	}
	if len(instances.TaskInstances) == 0 {
		t.Fatal("GetTaskInstances returned none")
	}
	for _, ti := range instances.TaskInstances {
		state := "<nil>"
		if ti.State != nil {
			state = *ti.State
		}
		t.Logf("task %s state=%s try=%d", ti.TaskID, state, ti.TryNumber)
	}

	// Airflow 3 returns structured log records; this must come back as text.
	logs, err := c.GetTaskLog(ctx, "monitor_dag", runID, "collect_failed_tasks", 1)
	if err != nil {
		t.Fatalf("GetTaskLog: %v", err)
	}
	if logs == "" {
		t.Error("GetTaskLog returned empty content")
	}
	t.Logf("log rendered to %d bytes", len(logs))

	entry, err := c.GetXComEntry(ctx, "monitor_dag", runID, "collect_failed_tasks", "return_value")
	if err != nil {
		t.Fatalf("GetXComEntry: %v", err)
	}
	if entry.Key != "return_value" {
		t.Errorf("xcom key = %q, want return_value", entry.Key)
	}
	t.Logf("xcom value type %T", entry.Value)
}

func TestLiveImportErrorsAndEventLogs(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	ie, err := c.GetImportErrors(ctx)
	if err != nil {
		t.Fatalf("GetImportErrors: %v", err)
	}
	t.Logf("import errors: %d", ie.TotalEntries)
	for _, e := range ie.ImportErrors {
		t.Errorf("unexpected import error in %s: %s", e.Filename, e.StackTrace)
	}

	logs, err := c.GetEventLogs(ctx, "monitor_dag", "", "", 100, 0)
	if err != nil {
		t.Fatalf("GetEventLogs: %v", err)
	}
	if logs.TotalEntries == 0 {
		t.Error("GetEventLogs returned nothing for monitor_dag")
	}
	t.Logf("event logs: %d (showing first)", logs.TotalEntries)
	if len(logs.EventLogs) > 0 {
		t.Logf("  %s: %s", logs.EventLogs[0].When.Format(time.RFC3339), logs.EventLogs[0].Event)
	}
}

func TestLiveConnectionsCRUD(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := liveCtx(t)
	defer cancel()

	const id = "cicada_live_test_conn"
	_ = c.DeleteConnection(ctx, id)

	host := "example.com"
	port := int32(1234)
	login := "user"
	extra := `{"k":"v"}`
	desc := "live test"

	created, err := c.CreateConnection(ctx, Connection{
		ConnectionID: id,
		ConnType:     "http",
		Description:  &desc,
		Host:         &host,
		Login:        &login,
		Port:         &port,
		Extra:        &extra,
	})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if created.ConnectionID != id {
		t.Errorf("created id = %q, want %q", created.ConnectionID, id)
	}
	defer func() { _ = c.DeleteConnection(ctx, id) }()

	got, err := c.GetConnection(ctx, id)
	if err != nil {
		t.Fatalf("GetConnection: %v", err)
	}
	if got.Host == nil || *got.Host != host {
		t.Errorf("host = %v, want %q", got.Host, host)
	}
	if got.Port == nil || *got.Port != port {
		t.Errorf("port = %v, want %d", got.Port, port)
	}

	newHost := "updated.example.com"
	updated, err := c.UpdateConnection(ctx, id, Connection{
		ConnectionID: id,
		ConnType:     "http",
		Host:         &newHost,
	})
	if err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}
	if updated.Host == nil || *updated.Host != newHost {
		t.Errorf("updated host = %v, want %q", updated.Host, newHost)
	}

	list, err := c.ListConnections(ctx, 100, 0, "connection_id")
	if err != nil {
		t.Fatalf("ListConnections: %v", err)
	}
	var found bool
	for _, conn := range list.Connections {
		if conn.ConnectionID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("ListConnections did not include %q", id)
	}

	if err := c.DeleteConnection(ctx, id); err != nil {
		t.Fatalf("DeleteConnection: %v", err)
	}
	if _, err := c.GetConnection(ctx, id); !IsNotFound(err) {
		t.Errorf("connection still present after delete; err = %v", err)
	}
}
