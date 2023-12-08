package airflow

import (
	"context"
	"crypto/tls"
	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"net/http"
	"strconv"
	"time"
)

type Connection struct {
	cli     *airflow.APIClient
	ctx     context.Context
	timeout time.Duration
}

var Conn *Connection

func Init() {
	conf := airflow.NewConfiguration()
	conf.Host = config.CMCicadaConfig.CMCicada.AirflowServer.Address
	useTLS, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.AirflowServer.UseTLS)
	if useTLS {
		skipTLSVerify, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.AirflowServer.SkipTLSVerify)
		conf.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipTLSVerify,
				},
			},
		}
		conf.Scheme = "https"
	} else {
		conf.Scheme = "http"
	}
	cli := airflow.NewAPIClient(conf)

	cred := airflow.BasicAuth{
		UserName: config.CMCicadaConfig.CMCicada.AirflowServer.Username,
		Password: config.CMCicadaConfig.CMCicada.AirflowServer.Password,
	}

	timeout, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.Timeout)
	ctx := context.WithValue(context.Background(), airflow.ContextBasicAuth, cred)
	conn := Connection{
		cli:     cli,
		ctx:     ctx,
		timeout: time.Duration(timeout) * time.Second,
	}

	Conn = &conn
}
