package common

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
)

func GetJSONRawBody(c echo.Context) (map[string]interface{}, error) {
	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		return nil, errors.New("empty json body")
	}

	return jsonBody, nil
}
