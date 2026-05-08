package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// CreateConnection godoc
//
//	@ID		create-connection
//	@Summary	Create Connection
//	@Description	Create an Airflow connection.
//	@Tags		[Connection]
//	@Accept		json
//	@Produce	json
//	@Param		request body 	model.Connection true "Connection content"
//	@Success	200	{object}	model.Connection	"Successfully created the connection."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to create connection."
//	@Router		/connection [post]
func CreateConnection(c echo.Context) error {
	var req model.Connection

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &req,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewConnectionService()
	created, err := svc.Create(&req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, created, " ")
}

// GetConnection godoc
//
//	@ID		get-connection
//	@Summary	Get Connection
//	@Description	Get an Airflow connection.
//	@Tags		[Connection]
//	@Accept		json
//	@Produce	json
//	@Param		connId path string true "Connection ID."
//	@Success	200	{object}	model.Connection	"Successfully retrieved the connection."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get connection."
//	@Router		/connection/{connId} [get]
func GetConnection(c echo.Context) error {
	connId, err := requireParam(c, "connId", "connId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewConnectionService()
	connection, err := svc.Get(connId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, connection, " ")
}

// ListConnection godoc
//
//	@ID		list-connection
//	@Summary	List Connections
//	@Description	List Airflow connections.
//	@Tags		[Connection]
//	@Accept		json
//	@Produce	json
//	@Param		page query string false "Page of the connection list."
//	@Param		row query string false "Row of the connection list."
//	@Param		orderBy query string false "Order by field, prefix with - to desc."
//	@Success	200	{object}	[]model.Connection	"Successfully retrieved the connection list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get connection list."
//	@Router		/connection [get]
func ListConnection(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	orderBy := c.QueryParam("orderBy")

	svc := service.NewConnectionService()
	connections, err := svc.List(page, row, orderBy)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, connections, " ")
}

// UpdateConnection godoc
//
//	@ID		update-connection
//	@Summary	Update Connection
//	@Description	Update an Airflow connection.
//	@Tags		[Connection]
//	@Accept		json
//	@Produce	json
//	@Param		connId path string true "Connection ID."
//	@Param		request body 	model.Connection true "Connection content"
//	@Success	200	{object}	model.Connection	"Successfully updated the connection."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to update connection."
//	@Router		/connection/{connId} [put]
func UpdateConnection(c echo.Context) error {
	connId, err := requireParam(c, "connId", "connId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var req model.Connection

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &req,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewConnectionService()
	updated, err := svc.Update(connId, &req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, updated, " ")
}

// DeleteConnection godoc
//
//	@ID		delete-connection
//	@Summary	Delete Connection
//	@Description	Delete an Airflow connection.
//	@Tags		[Connection]
//	@Accept		json
//	@Produce	json
//	@Param		connId path string true "Connection ID."
//	@Success	200	{object}	model.SimpleMsg	"Successfully deleted the connection."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to delete connection."
//	@Router		/connection/{connId} [delete]
func DeleteConnection(c echo.Context) error {
	connId, err := requireParam(c, "connId", "connId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewConnectionService()
	if err := svc.Delete(connId); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}
