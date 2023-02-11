package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type NoEscapeJSONSerializer struct{}

var _ echo.JSONSerializer = &NoEscapeJSONSerializer{}

func (d *NoEscapeJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (d *NoEscapeJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	// Does not escape <, >, and ?
	err := json.NewDecoder(c.Request().Body).Decode(i)
	var ute *json.UnmarshalTypeError
	var se *json.SyntaxError
	if ok := errors.As(err, &ute); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if ok := errors.As(err, &se); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}
