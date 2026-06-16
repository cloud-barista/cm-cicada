package model

// Infra/Node runtime info from CB-Tumblebug is parsed with the shared imdl
// cloud-model types (cloudmodel.InfraInfo / cloudmodel.NodeInfo); see lib/ssh.
// Only the SSH private key response has no imdl counterpart, so it stays local.

type TBSSHKeyInfo struct {
	PrivateKey string `json:"privateKey,omitempty"`
}
