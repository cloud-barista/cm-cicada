package airflow

import (
	"fmt"
	"github.com/cloud-barista/cm-cicada/model"
)

func CreateDAGFactoryYAML(DAG model.DAG) error {
	fmt.Println(DAG)
	return nil
}
