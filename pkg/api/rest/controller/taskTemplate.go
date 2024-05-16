package controller

import (
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// CreateTaskTemplate godoc
//
// @Summary		Create TaskTemplate
// @Description	Register the task template.
// @Tags		[Task Template]
// @Accept		json
// @Produce		json
// @Param		TaskTemplate body model.TaskTemplate true "task template of the node."
// @Success		200	{object}	model.TaskTemplate		"Successfully register the task template"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to register the task template"
// @Router			/task_template [post]
func CreateTaskTemplate(c echo.Context) error {
	return nil
}

// GetTaskTemplate godoc
//
// @Summary		Get TaskTemplate
// @Description	Get the task template.
// @Tags		[Task Template]
// @Accept		json
// @Produce		json
// @Param		uuid path string true "UUID of the TaskTemplate"
// @Success		200	{object}	model.TaskTemplate		"Successfully get the task template"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task template"
// @Router		/task_template/{uuid} [get]
func GetTaskTemplate(c echo.Context) error {
	return nil
}

// ListTaskTemplate godoc
//
// @Summary		List TaskTemplate
// @Description	Get a list of task template.
// @Tags		[Task Template]
// @Accept		json
// @Produce		json
// @Param		page query string false "Page of the task template list."
// @Param		row query string false "Row of the task template list."
// @Param		uuid query string false "UUID of the task template."
// @Param		name query string false "Task template name."
// @Success		200	{object}	[]model.TaskTemplate	"Successfully get a list of task template."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a list of task template."
// @Router			/task_template [get]
func ListTaskTemplate(c echo.Context) error {
	return nil
}

// UpdateTaskTemplate godoc
//
// @Summary		Update TaskTemplate
// @Description	Update the task template.
// @Tags		[Task Template]
// @Accept		json
// @Produce		json
// @Param		TaskTemplate body model.TaskTemplate true "task template to modify."
// @Success		200	{object}	model.TaskTemplate		"Successfully update the task template"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to update the task template"
// @Router		/task_template/{uuid} [put]
func UpdateTaskTemplate(c echo.Context) error {
	return nil
}

// DeleteTaskTemplate godoc
//
// @Summary		Delete TaskTemplate
// @Description	Delete the task template.
// @Tags		[Task Template]
// @Accept		json
// @Produce		json
// @Success		200	{object}	model.TaskTemplate		"Successfully delete the task template"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to delete the task template"
// @Router		/task_template/{uuid} [delete]
func DeleteTaskTemplate(c echo.Context) error {
	return nil
}
