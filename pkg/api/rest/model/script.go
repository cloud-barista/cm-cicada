package model

type RunScriptReq struct {
	NSID    string `json:"ns_id"`
	MCIID   string `json:"mci_id"`
	VMID    string `json:"vm_id"`
	Content string `json:"content"` // Base64 encoded script content.
}

type ScriptResult struct {
	IsSuccess bool   `json:"is_success"`
	Output    string `json:"output"`
	Error     string `json:"error"`
}

type SleepTimeReq struct {
	Time string `json:"time"` // Ex: 300 = 300sec, 1m 30s = 1min 30seconds, 1h 10m 15s = 1hour 10minutes 15seconds, 1d 1s = 1day 1second, No Input Default: 10s
}
