package controller

import (
	// "errors"
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	// "github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"net/http"
	"log"
)

// CreateTaskComponent godoc
//
// @Summary		Create TaskComponent
// @Description	Register the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		TaskComponent body model.TaskComponent true "task component of the node."
// @Success		200	{object}	model.TaskComponent		"Successfully register the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to register the task component"
// @Router			/task_component [post]
func CreateTaskComponent(c echo.Context) error {
	taskComponent := new(model.TaskComponent)
	log.Println("taskComponent ? ", taskComponent)
	err := c.Bind(taskComponent)
	if err != nil {
		return err
	}
	taskComponent, err = dao.TaskComponentCreate(taskComponent)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskComponent, " ")
}

// GetTaskComponent godoc
//
// @Summary		Get TaskComponent
// @Description	Get the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		id path string true "id of the TaskComponent"
// @Success		200	{object}	model.TaskComponent		"Successfully get the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task component"
// @Router		/task_component/{id} [get]
func GetTaskComponent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "id is empty")
	}
	taskComponent, err := dao.TaskComponentGet(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, taskComponent, "")
}

// ListTaskComponent godoc
//
// @Summary		List TaskComponent
// @Description	Get a list of task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		page query string false "Page of the task component list."
// @Param		row query string false "Row of the task component list."
// @Success		200	{object}	[]model.TaskComponent	"Successfully get a list of task component."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a list of task component."
// @Router			/task_component [get]
func ListTaskComponent(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskComponentList, err := dao.TaskComponentGetList(page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, taskComponentList, "")
}

// UpdateTaskComponent godoc
//
// @Summary		Update TaskComponent
// @Description	Update the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		TaskComponent body model.TaskComponent true "task component to modify."
// @Success		200	{object}	model.TaskComponent		"Successfully update the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to update the task component"
// @Router		/task_component/{uuid} [put]
func UpdateTaskComponent(c echo.Context) error {
	return nil
}

// DeleteTaskComponent godoc
//
// @Summary		Delete TaskComponent
// @Description	Delete the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Success		200	{object}	model.TaskComponent		"Successfully delete the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to delete the task component"
// @Router		/task_component/{uuid} [delete]
func DeleteTaskComponent(c echo.Context) error {
	return nil
}
