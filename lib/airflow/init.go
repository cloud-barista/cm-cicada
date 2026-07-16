package airflow

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cloud-barista/cm-cicada/lib/airflow/client"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/jollaman999/utils/logger"
)

// Client wraps the Airflow API client with cm-cicada's per-DAG serialization
// and model mapping.
type Client struct {
	api *client.Client
}

var airflowClient *Client

var initLock sync.Mutex

func GetClient() (*Client, error) {
	if airflowClient == nil {
		go func() {
			Init()
		}()
		return nil, fmt.Errorf("ERROR: Airflow client not initialized")
	}

	return airflowClient, nil
}

func timeout() time.Duration {
	seconds, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.Timeout)
	return time.Duration(seconds) * time.Second
}

func checkPing(api *client.Client) error {
	var err error
	var i int

	retry, _ := strconv.Atoi(config.CMCicadaConfig.CMCicada.AirflowServer.InitRetry)
	for i = 0; i < retry; i++ {
		logger.Println(logger.INFO, false, "Pinging Airflow Server... "+
			"(Trying: "+strconv.Itoa(i+1)+"/"+strconv.Itoa(retry)+")")

		ctx, cancel := context.WithTimeout(context.Background(), timeout())
		err = api.Health(ctx)
		cancel()

		if err == nil {
			break
		}
		time.Sleep(timeout())
	}

	if err != nil {
		return err
	}

	if i == retry {
		return errors.New("ERROR: Airflow Server is not responding")
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

// Context returns a request context bound to the configured timeout.
//
// Airflow 3 authenticates with a JWT that the client fetches and renews on its
// own, so unlike the Airflow 2 client there are no credentials in the context.
func Context() (context.Context, func()) {
	return context.WithTimeout(context.Background(), timeout())
}

func Init() {
	if !initLock.TryLock() {
		return
	}
	defer initLock.Unlock()

	useTLS, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.AirflowServer.UseTLS)
	skipTLSVerify, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.AirflowServer.SkipTLSVerify)

	api := client.New(client.Config{
		Address:       config.CMCicadaConfig.CMCicada.AirflowServer.Address,
		Username:      config.CMCicadaConfig.CMCicada.AirflowServer.Username,
		Password:      config.CMCicadaConfig.CMCicada.AirflowServer.Password,
		UseTLS:        useTLS,
		SkipTLSVerify: skipTLSVerify,
		Timeout:       timeout(),
	})

	err := checkPing(api)
	if err != nil {
		logger.Println(logger.ERROR, true, err.Error())
		return
	}

	airflowClient = &Client{api: api}

	registerConnections()
}
