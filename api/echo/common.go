package echo

import (
	"encoding/json"
	"errors"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"net/http"
)

func getJSONRawBody(c echo.Context) (map[string]interface{}, error) {
	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		return nil, errors.New("empty json body")
	}

	return jsonBody, nil
}

func returnErrorMsg(c echo.Context, msg string) error {
	return c.JSONPretty(http.StatusBadRequest, map[string]string{
		"error": msg,
	}, " ")
}

func returnInternalError(c echo.Context, err error, reason string) error {
	logger.Println(logger.ERROR, true, err.Error())

	return returnErrorMsg(c, "Internal error occurred. (Reason: "+reason+", Error: "+err.Error()+")")
}
