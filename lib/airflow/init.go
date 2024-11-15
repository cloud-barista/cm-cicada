package airflow

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/jollaman999/utils/logger"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	*airflow.APIClient
}

var airflowClient *Client

var initLock sync.Mutex

func GetClient() (*Client, error) {
	if airflowClient == nil {
		go func() {
			Init()
		}()
		return nil, fmt.Errorf("Airflow client not initialized. Try again later or check the Airflow server.")
	}

	return airflowClient, nil
}

func ping(url string) error {
	timeout, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.Timeout)
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	_, err := client.Get(url + "/health")
	return err
}

func checkPing(url string) error {
	var err error
	var i int
	timeout, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.Timeout)

	retry, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.InitRetry)
	for i = 0; i < retry; i++ {
		logger.Println(logger.INFO, false, "Pinging Airflow Server... "+
			"(Trying: "+strconv.Itoa(i+1)+"/"+strconv.Itoa(retry)+")")
		err = ping(url)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(timeout) * time.Second)
		}
	}

	if err != nil {
		return err
	}

	if i == retry {
		return errors.New("Airflow Server is not responding!")
	}

	return nil
}

func registerConnections() {
	for _, connection := range config.CMCicadaConfig.CMCicada.AirflowServer.Connections {
		logger.Println(logger.INFO, false, "Registering connection: ", connection)
		err := airflowClient.RegisterConnection(&connection)
		if err != nil {
			logger.Println(logger.ERROR, false, err.Error())
		}
	}
}

func Context() (context.Context, func()) {
	cred := airflow.BasicAuth{
		UserName: config.CMCicadaConfig.CMCicada.AirflowServer.Username,
		Password: config.CMCicadaConfig.CMCicada.AirflowServer.Password,
	}
	timeout, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.Timeout)

	return context.WithTimeout(context.WithValue(context.Background(), airflow.ContextBasicAuth, cred), time.Duration(timeout)*time.Second)
}

func Init() {
	if !initLock.TryLock() {
		return
	}
	defer initLock.Unlock()

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

	err := checkPing(conf.Scheme + "://" + conf.Host)
	if err != nil {
		logger.Println(logger.ERROR, true, err.Error())
		return
	}

	cli := airflow.NewAPIClient(conf)
	conn := Client{
		cli,
	}

	airflowClient = &conn

	registerConnections()
}
