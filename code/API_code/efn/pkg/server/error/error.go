/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package error

import (
	"net/http"
	"unicode"

	"github.com/labstack/echo/v4"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

// SendFromStatusCode sends a JSON response with a specified status code and error detail.
func SendFromStatusCode(c echo.Context, statusCode int, detail string) error {
	return c.JSON(statusCode, &models.ErrorInfo{
		Code:    http.StatusText(statusCode),
		Message: capitalizeFirst(detail),
		Status:  statusCode,
	})
}

// SendFromStatusCodeWithCode sends a JSON response allowing an explicit code string.
// Useful when you want to return standardized error codes instead of the default status text.
func SendFromStatusCodeWithCode(c echo.Context, statusCode int, code string, detail string) error {
	return c.JSON(statusCode, &models.ErrorInfo{
		Code:    code,
		Message: capitalizeFirst(detail),
		Status:  statusCode,
	})
}

// Send sends a JSON response based on an error.
// It determines the appropriate HTTP status code from the error.
func Send(c echo.Context, err error) error {
	detail := err.Error()
	statusCode := statusCodeFromError(err)
	return c.JSON(statusCode, models.ErrorInfo{
		Code:    http.StatusText(statusCode),
		Message: detail,
		Status:  statusCode,
	})
}

// statusCodeFromError maps specific errors to HTTP status codes.
// It uses errors.Is to properly detect wrapped errors.
func statusCodeFromError(err error) int {
	switch {
	// Convert your errors to http status code
	case IsNotFound(err):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
