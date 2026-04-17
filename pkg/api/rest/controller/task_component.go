package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

// CreateTaskComponent godoc
//
//	@ID				create-task-component
//	@Summary		Create Task Component
//	@Description	Register the task component.
//	@Tags		[Task Component]
//	@Accept		json
//	@Produce		json
//	@Param		TaskComponent body model.CreateTaskComponentReq true "task component to create."
//	@Success		200	{object}	model.TaskComponent		"Successfully register the task component"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to register the task component"
//	@Router		/task_component [post]
func CreateTaskComponent(c echo.Context) error {
	req := new(model.CreateTaskComponentReq)
	if err := c.Bind(req); err != nil {
		return err
	}

	svc := service.NewTaskComponentService()
	taskComponent, err := svc.Create(*req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskComponent, " ")
}

// GetTaskComponent godoc
//
//	@ID				get-task-component
//	@Summary		Get Task Component
//	@Description	Get the task component.
//	@Tags		[Task Component]
//	@Accept		json
//	@Produce		json
//	@Param		tcId path string true "ID of the TaskComponent"
//	@Success		200	{object}	model.TaskComponent		"Successfully get the task component"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task component"
//	@Router		/task_component/{tcId} [get]
func GetTaskComponent(c echo.Context) error {
	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "tcId is empty")
	}

	svc := service.NewTaskComponentService()
	taskComponent, err := svc.Get(tcId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskComponent, "")
}

// GetTaskComponentByName godoc
//
//	@ID				get-task-component-by-name
//	@Summary		Get Task Component by Name
//	@Description	Get the task component by name.
//	@Tags		[Task Component]
//	@Accept		json
//	@Produce		json
//	@Param		tcName path string true "Name of the TaskComponent"
//	@Success		200	{object}	model.TaskComponent		"Successfully get the task component"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task component"
//	@Router		/task_component/name/{tcName} [get]
func GetTaskComponentByName(c echo.Context) error {
	tcName := c.Param("tcName")
	if tcName == "" {
		return common.ReturnErrorMsg(c, "tcName is empty")
	}

	svc := service.NewTaskComponentService()
	taskComponent, err := svc.GetByName(tcName)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskComponent, "")
}

// ListTaskComponent godoc
//
//	@ID				list-task-component
//	@Summary		List Task Components
//	@Description	Get a list of task component.
//	@Tags			[Task Component]
//	@Accept			json
//	@Produce		json
//	@Param			page query string false "Page of the task component list."
//	@Param			row query string false "Row of the task component list."
//	@Success		200	{object}	[]model.TaskComponent	"Successfully get a list of task component."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get a list of task component."
//	@Router			/task_component [get]
func ListTaskComponent(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewTaskComponentService()
	taskComponentList, err := svc.List(page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskComponentList, "")
}

// UpdateTaskComponent godoc
//
//	@ID				update-task-component
//	@Summary		Update Task Component
//	@Description	Update the task component.
//	@Tags		[Task Component]
//	@Accept		json
//	@Produce		json
//	@Param		tcId path string true "ID of the TaskComponent"
//	@Param		TaskComponent body model.CreateTaskComponentReq true "task component to modify."
//	@Success		200	{object}	model.TaskComponent		"Successfully update the task component"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to update the task component"
//	@Router		/task_component/{tcId} [put]
func UpdateTaskComponent(c echo.Context) error {
	req := new(model.CreateTaskComponentReq)
	if err := c.Bind(req); err != nil {
		return err
	}

	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tcId.")
	}

	svc := service.NewTaskComponentService()
	updated, err := svc.Update(tcId, *req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, updated, " ")
}

// DeleteTaskComponent godoc
//
//	@ID				delete-task-component
//	@Summary		Delete Task Component
//	@Description	Delete the task component.
//	@Tags		[Task Component]
//	@Accept		json
//	@Produce		json
//	@Param		tcId path string true "ID of the task component."
//	@Success		200	{object}	model.SimpleMsg		"Successfully delete the task component"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to delete the task component"
//	@Router		/task_component/{tcId} [delete]
func DeleteTaskComponent(c echo.Context) error {
	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tcId.")
	}

	svc := service.NewTaskComponentService()
	if err := svc.Delete(tcId); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}
