package controller

import (
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func requireParam(c echo.Context, paramName, label string) (string, error) {
	value := c.Param(paramName)
	if value == "" {
		return "", errors.New("Please provide the " + label + ".")
	}
	return value, nil
}

func queryBool(c echo.Context, name string) (bool, error) {
	value := c.QueryParam(name)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

