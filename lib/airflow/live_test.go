package airflow

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// Exercises the lib/airflow wrapper (locking, model mapping) against a real
// Airflow 3 server. Skipped unless AIRFLOW_LIVE_TEST_ADDR is set:
//
//	AIRFLOW_LIVE_TEST_ADDR=127.0.0.1:8080 go test ./lib/airflow/ -run TestLive -v
//
// cmd/cm-cicada requires root, but the wrapper does not, so this covers the
// mapping layer without a privileged daemon.
func setupLiveClient(t *testing.T) *Client {
	t.Helper()

	addr := os.Getenv("AIRFLOW_LIVE_TEST_ADDR")
	if addr == "" {
		t.Skip("AIRFLOW_LIVE_TEST_ADDR not set; skipping live Airflow test")
	}

	cfg := &config.CMCicadaConfig.CMCicada.AirflowServer
	cfg.Address = addr
	cfg.UseTLS = "false"
	cfg.SkipTLSVerify = "false"
	cfg.InitRetry = "3"
	cfg.Timeout = "30"
	if cfg.Username = os.Getenv("AIRFLOW_LIVE_TEST_USER"); cfg.Username == "" {
		cfg.Username = "airflow"
	}
	if cfg.Password = os.Getenv("AIRFLOW_LIVE_TEST_PASSWORD"); cfg.Password == "" {
		cfg.Password = "airflow_pass"
	}
	// Init() registers these; keep it empty so the test controls its own.
	cfg.Connections = nil

	airflowClient = nil
	Init()

	client, err := GetClient()
	if err != nil {
		t.Fatalf("GetClient after Init: %v", err)
	}
	return client
}

func TestLiveWrapperDAGLookup(t *testing.T) {
	client := setupLiveClient(t)

	dag, err := client.GetDAG("monitor_dag")
	if err != nil {
		t.Fatalf("GetDAG: %v", err)
	}
	if dag.DagID != "monitor_dag" {
		t.Errorf("DagID = %q, want monitor_dag", dag.DagID)
	}
}

func TestLiveWrapperRunDAG(t *testing.T) {
	client := setupLiveClient(t)

	run, err := client.RunDAG("templates", map[string]interface{}{"from": "wrapper-test"})
	if err != nil {
		t.Fatalf("RunDAG: %v", err)
	}
	if run.DagRunID == "" {
		t.Error("RunDAG returned an empty dag_run_id")
	}
	t.Logf("RunDAG -> %s (state=%s)", run.DagRunID, run.State)

	// A missing DAG must surface as this exact message rather than a transport
	// error; RunDAG checks existence before triggering.
	_, err = client.RunDAG("definitely_missing_dag", nil)
	if err == nil {
		t.Fatal("RunDAG on a missing DAG returned no error")
	}
	if err.Error() != "provided dag_id is not exist" {
		t.Errorf("RunDAG missing-DAG error = %q, want %q", err.Error(), "provided dag_id is not exist")
	}
}

func TestLiveWrapperRunsAndStatus(t *testing.T) {
	client := setupLiveClient(t)

	runs, err := client.GetDAGRuns("monitor_dag")
	if err != nil {
		t.Fatalf("GetDAGRuns: %v", err)
	}
	if len(runs.DagRuns) == 0 {
		t.Skip("no monitor_dag runs available")
	}

	// Every state the wrapper advertises must be accepted by the server.
	for _, state := range client.GetAllowedDagStateEnumValues() {
		if _, err := client.GetDagStatus("monitor_dag", state); err != nil {
			t.Errorf("GetDagStatus(%s): %v", state, err)
		}
	}

	runID := runs.DagRuns[len(runs.DagRuns)-1].DagRunID
	instances, err := client.GetTaskInstances("monitor_dag", runID)
	if err != nil {
		t.Fatalf("GetTaskInstances: %v", err)
	}
	if len(instances.TaskInstances) == 0 {
		t.Fatal("GetTaskInstances returned none")
	}

	logs, err := client.GetTaskLogs("monitor_dag", runID, "collect_failed_tasks", 1)
	if err != nil {
		t.Fatalf("GetTaskLogs: %v", err)
	}
	if logs == "" {
		t.Error("GetTaskLogs returned empty text")
	}

	// collect_failed_tasks returns a dict, which Airflow 3 hands back already
	// decoded rather than as the JSON string the Airflow 2 API returned.
	xcom, err := client.GetXComValue("monitor_dag", runID, "collect_failed_tasks", "return_value")
	if err != nil {
		t.Fatalf("GetXComValue: %v", err)
	}
	if xcom == nil {
		t.Fatal("GetXComValue returned nil")
	}
	if _, ok := xcom["dag_state"]; !ok {
		t.Errorf("xcom map missing dag_state key: %#v", xcom)
	}
	t.Logf("xcom dag_state=%v failed_tasks=%v", xcom["dag_state"], xcom["failed_tasks"])
}

func TestLiveWrapperImportErrors(t *testing.T) {
	client := setupLiveClient(t)

	result, err := client.GetImportErrors()
	if err != nil {
		t.Fatalf("GetImportErrors: %v", err)
	}
	for _, e := range result.ImportErrors {
		t.Errorf("unexpected import error in %s: %s", e.Filename, e.StackTrace)
	}
	t.Logf("import errors: %d", result.TotalEntries)
}

func TestLiveWrapperEventLogs(t *testing.T) {
	client := setupLiveClient(t)

	logs, err := client.GetEventLogs("monitor_dag", "", "")
	if err != nil {
		t.Fatalf("GetEventLogs: %v", err)
	}
	if logs.TotalEntries == 0 {
		t.Error("GetEventLogs returned nothing")
	}
	t.Logf("event logs: %d", logs.TotalEntries)
}

// Covers RegisterConnection and the model.Connection <-> client.Connection
// mapping, which is what cm-cicada does on startup.
func TestLiveWrapperConnectionMapping(t *testing.T) {
	client := setupLiveClient(t)

	conn := &model.Connection{
		ID:          "cicada_wrapper_test_conn",
		Type:        "http",
		Description: "wrapper test",
		Host:        "api.example.com",
		Port:        9999,
		Schema:      "https",
		Login:       "someone",
		Password:    "secret",
		Extra:       `{"timeout":5}`,
	}

	deleted := false
	defer func() {
		if !deleted {
			_ = client.DeleteConnection(conn.ID)
		}
	}()

	// RegisterConnection must be idempotent: Init() re-runs it every start.
	for i := 0; i < 2; i++ {
		if err := client.RegisterConnection(conn); err != nil {
			t.Fatalf("RegisterConnection (attempt %d): %v", i+1, err)
		}
	}

	got, err := client.GetConnection(conn.ID)
	if err != nil {
		t.Fatalf("GetConnection: %v", err)
	}
	if got.ID != conn.ID || got.Type != conn.Type {
		t.Errorf("id/type = %q/%q, want %q/%q", got.ID, got.Type, conn.ID, conn.Type)
	}
	if got.Host != conn.Host {
		t.Errorf("host = %q, want %q", got.Host, conn.Host)
	}
	if got.Port != conn.Port {
		t.Errorf("port = %d, want %d", got.Port, conn.Port)
	}
	if got.Schema != conn.Schema {
		t.Errorf("schema = %q, want %q", got.Schema, conn.Schema)
	}
	if got.Login != conn.Login {
		t.Errorf("login = %q, want %q", got.Login, conn.Login)
	}
	// Airflow parses and re-serializes extra, so it comes back reformatted
	// ({"timeout": 5}). Compare the decoded value, not the bytes.
	if !sameJSON(t, got.Extra, conn.Extra) {
		t.Errorf("extra = %q, want something equivalent to %q", got.Extra, conn.Extra)
	}

	list, err := client.ListConnections(100, 0, "connection_id")
	if err != nil {
		t.Fatalf("ListConnections: %v", err)
	}
	var found bool
	for _, c := range list {
		if c.ID == conn.ID {
			found = true
			if c.Host != conn.Host {
				t.Errorf("listed host = %q, want %q", c.Host, conn.Host)
			}
			// Airflow 3 returns the full connection in listings and only masks
			// extra keys it recognises as secret, so a custom-named credential
			// would leak. Listings must stay narrow.
			if c.Extra != "" {
				t.Errorf("ListConnections exposed extra %q; listings must not carry credentials", c.Extra)
			}
			if c.Password != "" {
				t.Errorf("ListConnections exposed password %q; listings must not carry credentials", c.Password)
			}
		}
	}
	if !found {
		t.Errorf("ListConnections did not include %q", conn.ID)
	}

	if err := client.DeleteConnection(conn.ID); err != nil {
		t.Fatalf("DeleteConnection: %v", err)
	}
	deleted = true

	if _, err := client.GetConnection(conn.ID); err == nil {
		t.Error("connection is still readable after delete")
	}
}

// sameJSON reports whether two JSON documents are semantically equal.
func sameJSON(t *testing.T, a, b string) bool {
	t.Helper()
	var av, bv interface{}
	if err := json.Unmarshal([]byte(a), &av); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bv); err != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}
