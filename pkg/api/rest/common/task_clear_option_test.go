package common

import (
	"testing"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/mitchellh/mapstructure"
)

// decodeTaskClearOption decodes a request body the way the ClearTaskInstances
// controller does (JSON object -> map -> mapstructure).
func decodeTaskClearOption(t *testing.T, body map[string]interface{}) model.TaskClearOption {
	t.Helper()
	var opt model.TaskClearOption
	if err := mapstructure.Decode(body, &opt); err != nil {
		t.Fatalf("decode %v: %v", body, err)
	}
	return opt
}

func TestTaskClearOptionResetDagRuns(t *testing.T) {
	tests := []struct {
		name      string
		body      map[string]interface{}
		wantReset bool
		wantErr   bool
	}{
		{"omitted on real clear follows Airflow default", map[string]interface{}{"dryRun": false}, true, false},
		{"omitted on dry run", map[string]interface{}{"dryRun": true}, false, false},
		{"explicit false is kept", map[string]interface{}{"dryRun": false, "resetDagRuns": false}, false, false},
		{"explicit true is kept", map[string]interface{}{"dryRun": false, "resetDagRuns": true}, true, false},
		{"explicit false on dry run", map[string]interface{}{"dryRun": true, "resetDagRuns": false}, false, false},
		{"explicit true on dry run is rejected", map[string]interface{}{"dryRun": true, "resetDagRuns": true}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{"taskIds": []interface{}{"t1"}, "includeDownstream": true}
			for k, v := range tt.body {
				body[k] = v
			}
			opt := decodeTaskClearOption(t, body)

			err := ValidateTaskClearOptions(opt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTaskClearOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got := opt.ResetDagRunsOrDefault(); got != tt.wantReset {
				t.Fatalf("ResetDagRunsOrDefault() = %v, want %v", got, tt.wantReset)
			}
		})
	}
}
