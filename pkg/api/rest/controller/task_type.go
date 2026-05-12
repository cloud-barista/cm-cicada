package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/lib/airflow/catalog"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ catalog.TaskTypeDef // swag type reference

// ListTaskType godoc
//
//	@ID				list-task-type
//	@Summary		List Task Types
//	@Description	List all available task types defined in the catalog.
//	@Tags			[Task Type]
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		catalog.TaskTypeDef		"Successfully list task types"
//	@Failure		500	{object}	common.ErrorResponse	"Failed to list task types"
//	@Router			/task_type [get]
func ListTaskType(c echo.Context) error {
	svc := service.NewTaskTypeService()
	return c.JSONPretty(http.StatusOK, svc.List(), " ")
}

// GetTaskType godoc
//
//	@ID				get-task-type
//	@Summary		Get Task Type
//	@Description	Get a specific task type definition (including schemas).
//	@Tags			[Task Type]
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Task type id (e.g. http, bash, ssh)"
//	@Success		200	{object}	catalog.TaskTypeDef		"Successfully get the task type"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		404	{object}	common.ErrorResponse	"Task type not found"
//	@Router			/task_type/{id} [get]
func GetTaskType(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "id is empty")
	}

	svc := service.NewTaskTypeService()
	def, err := svc.Get(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, def, " ")
}
