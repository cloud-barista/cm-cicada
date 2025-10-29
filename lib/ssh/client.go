package ssh

import (
	"encoding/json"
	"errors"
	"strings"

	"net"
	"strconv"
	"time"

	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	*goph.Client
	SSHTarget *model.SSHTarget
	nsID      string
	mciID     string
	id        string
}

func (c *Client) NewSessionWithRetry() (*ssh.Session, error) {
	var session *ssh.Session
	var err error

	// Try to create session with existing connection first
	for retry := 0; retry < 3; retry++ {
		session, err = c.NewSession()
		if err == nil {
			return session, nil
		}

		// If EOF error, try to reconnect
		if err.Error() == "EOF" {
			// Close existing connection
			if c.Client != nil {
				_ = c.Close()
			}

			// Recreate connection
			newClient, reconnectErr := NewSSHClient(c.nsID, c.mciID, c.id)
			if reconnectErr != nil {
				time.Sleep(time.Second * 2)
				continue
			}

			// Update client
			c.Client = newClient.Client
			continue
		}

		// For other errors, just retry after delay
		time.Sleep(time.Second * 2)
	}

	return nil, err
}

func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey) error {
	hostFound, _ := goph.CheckKnownHost(host, remote, key, "")

	if hostFound {
		return nil
	}

	return goph.AddKnownHost(host, remote, key, "")
}

func NewSSHClient(nsID string, mciID string, vmID string) (*Client, error) {
	var client *goph.Client
	var sshTarget *model.SSHTarget

	var connection model.Connection
	for _, connection = range config.CMCicadaConfig.CMCicada.AirflowServer.Connections {
		if strings.Contains(strings.ToLower(connection.ID), "tumblebug") {
			break
		}
	}

	data, err := common.GetHTTPRequest("http://"+connection.Host+
		":"+strconv.Itoa(int(connection.Port))+
		"/tumblebug/ns/"+nsID+"/mci/"+mciID+"/vm/"+vmID,
		connection.Login, connection.Password)
	if err != nil {
		return nil, err
	}

	var vmInfo model.TBVMInfo
	err = json.Unmarshal(data, &vmInfo)
	if err != nil {
		return nil, err
	}

	sshPort, err := strconv.Atoi(vmInfo.SSHPort)
	if err != nil {
		return nil, errors.New("invalid ssh port")
	}

	data, err = common.GetHTTPRequest("http://"+connection.Host+
		":"+strconv.Itoa(int(connection.Port))+
		"/tumblebug/ns/"+nsID+"/resources/sshKey/"+vmInfo.SSHKeyID,
		connection.Login, connection.Password)
	if err != nil {
		return nil, err
	}

	var sshKeyInfo model.TBSSHKeyInfo
	err = json.Unmarshal(data, &sshKeyInfo)
	if err != nil {
		return nil, err
	}

	if sshKeyInfo.PrivateKey == "" {
		return nil, errors.New("failed to get private key")
	}

	var auth goph.Auth
	auth, err = goph.RawKey(sshKeyInfo.PrivateKey, "")
	if err != nil {
		return nil, err
	}

	client, err = goph.NewConn(&goph.Config{
		User:     vmInfo.VMUserName,
		Addr:     vmInfo.PublicIP,
		Port:     uint(sshPort),
		Auth:     auth,
		Timeout:  goph.DefaultTimeout,
		Callback: AddKnownHost,
	})
	if err != nil {
		return nil, err
	}

	sshTarget = &model.SSHTarget{
		IP:         vmInfo.PublicIP,
		Port:       uint(sshPort),
		UseKeypair: true,
		Username:   vmInfo.VMUserName,
		Password:   "",
		PrivateKey: sshKeyInfo.PrivateKey,
	}

	return &Client{
		Client:    client,
		SSHTarget: sshTarget,
		nsID:      nsID,
		mciID:     mciID,
		id:        vmID,
	}, nil
}
