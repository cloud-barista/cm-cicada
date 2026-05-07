package controller

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/common" // Need for swag
)

// ExampleSampleData returns a deterministic JSON payload for use as the
// upstream task in xcom-based workflow examples. The shape mirrors the body
// expected by ExampleEcho so a workflow can pipe one into the other without
// transformation.
//
//	@ID				example-get-data
//	@Summary		Sample data for xcom example
//	@Description	Return a deterministic JSON payload used by the http_xcom workflow example.
//	@Tags			[Example]
//	@Produce		json
//	@Success		200 {object} map[string]any "Sample payload"
//	@Router			/example/data [get]
func ExampleSampleData(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, map[string]any{
		"id":      "example-001",
		"name":    "sample-record",
		"payload": map[string]any{"value": 42, "tag": "demo"},
	}, " ")
}

// ExampleEcho echoes the raw request body back as the response. Used as the
// downstream task in xcom-based workflow examples to verify the upstream body
// arrived intact.
//
//	@ID				example-post-echo
//	@Summary		Echo request body
//	@Description	Echo the raw request body back. Used by the http_xcom workflow example to verify body propagation.
//	@Tags			[Example]
//	@Accept			json
//	@Produce		json
//	@Success		200 {object} map[string]any "Echoed body"
//	@Router			/example/echo [post]
func ExampleEcho(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, body)
}
