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
	if req.MCIID == "" {
		return nil, errors.New("please provide the mci_id")
	}
	if req.VMID == "" {
		return nil, errors.New("please provide the vm_id")
	}
	if req.Content == "" {
		return nil, errors.New("please provide the content")
	}

	var result model.ScriptResult
	output, err := ssh.ExecuteScript(req.NSID, req.MCIID, req.VMID, req.Content)
	if err != nil {
		result.IsSuccess = false
		result.Error = err.Error()
	} else {
		result.IsSuccess = true
	}
	result.Output = string(output)

	return &result, nil
}

// validSleepDuration accepts Go-style durations like "10s", "1m30s", "500ms"
// or plain integer seconds like "10". Rejects anything else to prevent
// shell injection via cmd.RunBash.
var validSleepDuration = regexp.MustCompile(`^[0-9]+([hms]|ms)?$`)

func (s *CicadaBuiltinService) SleepTime(req model.SleepTimeReq) (*model.SimpleMsg, error) {
	duration := req.Time
	if duration == "" {
		duration = "10s"
	}

	if !validSleepDuration.MatchString(duration) {
		return nil, errors.New("invalid time format: must be a number followed by optional h/m/s/ms (e.g. 10s, 1m, 500ms)")
	}

	_, err := cmd.RunBash("sleep " + duration)
	if err != nil {
		return &model.SimpleMsg{Message: err.Error()}, nil
	}

	return &model.SimpleMsg{Message: "success"}, nil
}
