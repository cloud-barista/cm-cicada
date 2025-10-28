package ssh

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/jollaman999/utils/logger"
)

func decodeScript(base64EncodedContent string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64EncodedContent)
	if err != nil {
		return "", fmt.Errorf("failed to decode script content %s: %v", base64EncodedContent, err)
	}

	script := string(decoded)
	script = strings.ReplaceAll(script, "\r\n", "\n")
	script = strings.ReplaceAll(script, "\r", "\n")

	return script, nil
}

func ExecuteScript(nsID string, mciID string, vmID string, base64EncodedContent string) ([]byte, error) {
	var targetClient *Client

	targetClient, err := NewSSHClient(nsID, mciID, vmID)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to connect to target host: NS_ID: %s, MCI_ID: %s, VM_ID: %s (Error: %v)",
			nsID, mciID, vmID, err)
	}

	defer func() {
		_ = targetClient.Close()
	}()

	script, err := decodeScript(base64EncodedContent)
	if err != nil {
		return []byte{}, err
	}

	session, err := targetClient.NewSessionWithRetry()
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		_ = session.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				keepAliveSession, err := targetClient.NewSessionWithRetry()
				if err != nil {
					logger.Printf(logger.ERROR, true, "Keep-alive session creation failed for target: "+
						"NS_ID: %s, MCI_ID: %s, VM_ID:%s\n", targetClient.nsID, targetClient.mciID, targetClient.id)
					continue
				}

				_, _ = keepAliveSession.CombinedOutput("echo keepalive")
				_ = keepAliveSession.Close()
			}
		}
	}()

	cmd := fmt.Sprintf("cat << 'SCRIPT_EOF' | bash\n%s\nSCRIPT_EOF", script)

	logger.Printf(logger.DEBUG, true, "Executing script with keep-alive enabled for target: "+
		"NS_ID: %s, MCI_ID: %s, VM_ID:%s\n", targetClient.nsID, targetClient.mciID, targetClient.id)
	output, err := session.CombinedOutput(cmd)

	cancel()

	return output, err
}
