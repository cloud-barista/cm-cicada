package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/cloud-barista/cm-cicada/db"
	restcommon "github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunWorkflowReturnsBadRequestWhenWorkflowDoesNotExist(t *testing.T) {
	setupWorkflowRunControllerTestDB(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/workflow/missing-id/run", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/workflow/:wfId/run")
	ctx.SetParamNames("wfId")
	ctx.SetParamValues("missing-id")

	if err := RunWorkflow(ctx); err != nil {
		t.Fatalf("RunWorkflow() returned unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("RunWorkflow() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var response restcommon.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "workflow not found with the provided id" {
		t.Fatalf("RunWorkflow() error message = %q, want %q", response.Error, "workflow not found with the provided id")
	}
}

func setupWorkflowRunControllerTestDB(t *testing.T) {
	t.Helper()

	oldDB := db.DB
	testDBPath := filepath.Join(t.TempDir(), "controller-test.db")
	testDB, err := gorm.Open(sqlite.Open(testDBPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(&model.Workflow{}); err != nil {
		t.Fatalf("failed to migrate workflow table: %v", err)
	}

	db.DB = testDB

	t.Cleanup(func() {
		if db.DB != nil {
			sqlDB, err := db.DB.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}
		db.DB = oldDB
	})
}
