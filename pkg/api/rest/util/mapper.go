package util

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
)

// CreateDataReqToData converts CreateDataReq to Data based on the specification version
func CreateDataReqToData(specVersion string, createDataReq model.CreateDataReq) (model.Data, error) {
	specVersionSpilit := strings.Split(specVersion, ".")
	if len(specVersionSpilit) != 2 {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMajor, err := strconv.Atoi(specVersionSpilit[0])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMinor, err := strconv.Atoi(specVersionSpilit[1])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	var taskGroups []model.TaskGroup
	var allTasks []model.Task

	if specVersionMajor > 0 && specVersionMajor <= 1 {
		if specVersionMinor == 0 {
			// v1.0
			for _, tgReq := range createDataReq.TaskGroups {
				var tasks []model.Task
				for _, tReq := range tgReq.Tasks {
					tasks = append(tasks, model.Task{
						ID:            uuid.New().String(),
						Name:          tReq.Name,
						TaskComponent: tReq.TaskComponent,
						RequestBody:   tReq.RequestBody,
						PathParams:    tReq.PathParams,
						QueryParams:   tReq.QueryParams,
						Dependencies:  tReq.Dependencies,
					})
				}

				allTasks = append(allTasks, tasks...)
				taskGroups = append(taskGroups, model.TaskGroup{
					ID:          uuid.New().String(),
					Name:        tgReq.Name,
					Description: tgReq.Description,
					Tasks:       tasks,
				})
			}

			for i, tgReq := range createDataReq.TaskGroups {
				for j, tg := range taskGroups {
					if tgReq.Name == tg.Name {
						if i == j {
							continue
						}

						return model.Data{}, errors.New("Duplicated task group name: " + tg.Name)
					}
				}
			}

			for i, tCheck := range allTasks {
				for j, t := range allTasks {
					if tCheck.Name == t.Name {
						if i == j {
							continue
						}

						return model.Data{}, errors.New("Duplicated task name: " + t.Name)
					}
				}
			}
		} else {
			return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
		}
	} else {
		return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
	}

	return model.Data{
		Description: createDataReq.Description,
		TaskGroups:  taskGroups,
	}, nil
}
