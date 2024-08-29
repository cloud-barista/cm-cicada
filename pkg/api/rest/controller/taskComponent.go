package controller

import (
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

// CreateTaskComponent godoc
//
// @Summary		Create TaskComponent
// @Description	Register the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		TaskComponent body model.CreateTaskComponentReq true "task component of the node."
// @Success		200	{object}	model.TaskComponent		"Successfully register the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to register the task component"
// @Router		/task_component [post]
func CreateTaskComponent(c echo.Context) error {
	createTaskComponentReq := new(model.CreateTaskComponentReq)
	err := c.Bind(createTaskComponentReq)
	if err != nil {
		return err
	}

	if createTaskComponentReq.Name == "" {
		return common.ReturnErrorMsg(c, "Please provide the name.")
	}

	var taskComponent model.TaskComponent
	taskComponent.ID = uuid.New().String()
	taskComponent.Name = createTaskComponentReq.Name
	taskComponent.Data = createTaskComponentReq.Data

	_, err = dao.TaskComponentCreate(&taskComponent)
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
// @Param		tcId path string true "ID of the TaskComponent"
// @Success		200	{object}	model.TaskComponent		"Successfully get the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task component"
// @Router		/task_component/{tcId} [get]
func GetTaskComponent(c echo.Context) error {
	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "tcId is empty")
	}
	taskComponent, err := dao.TaskComponentGet(tcId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, taskComponent, "")
}

// GetTaskComponentByName godoc
//
// @Summary		Get TaskComponent by Name
// @Description	Get the task component by name.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		tcName path string true "Name of the TaskComponent"
// @Success		200	{object}	model.TaskComponent		"Successfully get the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task component"
// @Router		/task_component/name/{tcName} [get]
func GetTaskComponentByName(c echo.Context) error {
	tcName := c.Param("tcName")
	if tcName == "" {
		return common.ReturnErrorMsg(c, "tcName is empty")
	}
	taskComponent := db.TaskComponentGetByName(tcName)
	if taskComponent == nil {
		return common.ReturnErrorMsg(c, "task component not found with the provided name")
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
// @Router		/task_component [get]
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
// @Param		tcId path string true "ID of the TaskComponent"
// @Param		TaskComponent body model.CreateTaskComponentReq true "task component to modify."
// @Success		200	{object}	model.TaskComponent		"Successfully update the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to update the task component"
// @Router		/task_component/{tcId} [put]
func UpdateTaskComponent(c echo.Context) error {
	taskComponent := new(model.CreateTaskComponentReq)
	err := c.Bind(taskComponent)
	if err != nil {
		return err
	}

	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tcId.")
	}
	oldTaskComponent, err := dao.TaskComponentGet(tcId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if taskComponent.Name != "" {
		oldTaskComponent.Name = taskComponent.Name
	}

	oldTaskComponent.Data = taskComponent.Data

	err = dao.TaskComponentUpdate(oldTaskComponent)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, oldTaskComponent, " ")
}

// DeleteTaskComponent godoc
//
// @Summary		Delete TaskComponent
// @Description	Delete the task component.
// @Tags		[Task Component]
// @Accept		json
// @Produce		json
// @Param		tcId path string true "ID of the task component."
// @Success		200	{object}	model.SimpleMsg		"Successfully delete the task component"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to delete the task component"
// @Router		/task_component/{tcId} [delete]
func DeleteTaskComponent(c echo.Context) error {
	tcId := c.Param("tcId")
	if tcId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tcId.")
	}

	taskComponent, err := dao.TaskComponentGet(tcId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = dao.TaskComponentDelete(taskComponent)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}
