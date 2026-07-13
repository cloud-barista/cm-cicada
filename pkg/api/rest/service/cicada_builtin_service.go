package service

import (
	"errors"
	"regexp"

	"github.com/cloud-barista/cm-cicada/lib/cmd"
	"github.com/cloud-barista/cm-cicada/lib/ssh"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

type CicadaBuiltinService struct{}

func NewCicadaBuiltinService() *CicadaBuiltinService {
	return &CicadaBuiltinService{}
}

func (s *CicadaBuiltinService) RunScript(req model.RunScriptReq) (*model.ScriptResult, error) {
	if req.NSID == "" {
		return nil, errors.New("please provide the ns_id")
	}
	if req.InfraID == "" {
		return nil, errors.New("please provide the infra_id")
	}
	if req.NodeID == "" {
		return nil, errors.New("please provide the node_id")
	}
	if req.Content == "" {
		return nil, errors.New("please provide the content")
	}

	var result model.ScriptResult
	output, err := ssh.ExecuteScript(req.NSID, req.InfraID, req.NodeID, req.Content)
	if err != nil {
		result.IsSuccess = false
		result.Error = err.Error()
	} else {
		result.IsSuccess = true
	}
	result.Output = string(output)

	return &result, nil
}

// validSleepDuration accepts one or more space-separated components, each a
// number with an optional unit (ms/s/m/h/d) — e.g. "10", "1m", "500ms",
// "1m 30s", "1h 10m 15s". `sleep` sums space-separated durations, so the
// compound forms work as written. Only digits, unit letters and single spaces
// are allowed, which keeps the value safe to pass to cmd.RunBash (no shell
// metacharacters).
var validSleepDuration = regexp.MustCompile(`^[0-9]+(ms|[smhd])?( [0-9]+(ms|[smhd])?)*$`)

func (s *CicadaBuiltinService) SleepTime(req model.SleepTimeReq) (*model.SimpleMsg, error) {
	duration := req.Time
	if duration == "" {
		duration = "10s"
	}

	if !validSleepDuration.MatchString(duration) {
		return nil, errors.New("invalid time format: numbers with optional ms/s/m/h/d units, optionally space-separated (e.g. 10s, 1m, 500ms, 1m 30s)")
	}

	_, err := cmd.RunBash("sleep " + duration)
	if err != nil {
		return &model.SimpleMsg{Message: err.Error()}, nil
	}

	return &model.SimpleMsg{Message: "success"}, nil
}
