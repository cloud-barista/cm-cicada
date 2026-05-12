package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task_types.yaml")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return path
}

func TestLoad_Success(t *testing.T) {
	path := writeTempYAML(t, `
task_types:
  - id: http
    label: HTTP
    category: api
    operator_class: foo.SimpleHttpOperator
    component_schema:
      endpoint: { type: string, required: true }
  - id: bash
    label: Bash
    category: utility
    operator_class: foo.BashOperator
    task_schema:
      bash_command: { type: string, required: true }
`)
	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	got, ok := Get("http")
	if !ok {
		t.Fatal("Get(http) ok=false")
	}
	if got.OperatorClass != "foo.SimpleHttpOperator" {
		t.Errorf("operator_class = %q, want foo.SimpleHttpOperator", got.OperatorClass)
	}
	if got.ComponentSchema["endpoint"].Type != "string" {
		t.Errorf("endpoint.Type = %q, want string", got.ComponentSchema["endpoint"].Type)
	}
	if !got.ComponentSchema["endpoint"].Required {
		t.Error("endpoint.Required = false, want true")
	}

	if !Has("bash") {
		t.Error("Has(bash) = false, want true")
	}
	if Has("nonexistent") {
		t.Error("Has(nonexistent) = true, want false")
	}

	list := List()
	if len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
	if list[0].ID != "http" || list[1].ID != "bash" {
		t.Errorf("List order = [%s, %s], want [http, bash]", list[0].ID, list[1].ID)
	}
}

func TestLoad_DuplicateID(t *testing.T) {
	path := writeTempYAML(t, `
task_types:
  - id: http
    operator_class: foo.X
  - id: http
    operator_class: foo.Y
`)
	if err := Load(path); err == nil {
		t.Error("Load expected error for duplicate id")
	}
}

func TestLoad_EmptyOperatorClass(t *testing.T) {
	path := writeTempYAML(t, `task_types: [ { id: x, operator_class: "" } ]`)
	if err := Load(path); err == nil {
		t.Error("Load expected error for empty operator_class")
	}
}

func TestLoad_EmptyID(t *testing.T) {
	path := writeTempYAML(t, `task_types: [ { id: "", operator_class: foo.X } ]`)
	if err := Load(path); err == nil {
		t.Error("Load expected error for empty id")
	}
}

func TestLoad_EmptyCatalog(t *testing.T) {
	path := writeTempYAML(t, `task_types: []`)
	if err := Load(path); err == nil {
		t.Error("Load expected error for empty task_types")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if err := Load("/nonexistent/path.yaml"); err == nil {
		t.Error("Load expected error for missing file")
	}
}
