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
