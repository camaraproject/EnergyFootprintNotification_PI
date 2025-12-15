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
package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	. "github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Retrieves the overall carbon footprint for the target Application instances in a certain period of time.
	// (POST /calculate-carbon-footprint)
	CalculateCarbonFootprint(ctx echo.Context, params CalculateCarbonFootprintParams) error
	// Provides the overall Energy Consumption for the target Application instances in a certain period of time.
	// (POST /calculate-energy-consumption)
	CalculateEnergyConsumption(ctx echo.Context, params CalculateEnergyConsumptionParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// CalculateCarbonFootprint converts echo context to params.
func (w *ServerInterfaceWrapper) CalculateCarbonFootprint(ctx echo.Context) error {
	var err error

	ctx.Set(OpenIdScopes, []string{"energy-footprint-notification:calculate-carbon-footprint"})

	// Parameter object where we will unmarshal all parameters from the context
	var params CalculateCarbonFootprintParams

	headers := ctx.Request().Header
	// ------------- Optional header parameter "x-correlator" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("x-correlator")]; found {
		var XCorrelator XCorrelator
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for x-correlator, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "x-correlator", valueList[0], &XCorrelator, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: false})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter x-correlator: %s", err))
		}

		params.XCorrelator = &XCorrelator
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.CalculateCarbonFootprint(ctx, params)
	return err
}

// CalculateEnergyConsumption converts echo context to params.
func (w *ServerInterfaceWrapper) CalculateEnergyConsumption(ctx echo.Context) error {
	var err error

	ctx.Set(OpenIdScopes, []string{"energy-footprint-notification:calculate-energy-consumption"})

	// Parameter object where we will unmarshal all parameters from the context
	var params CalculateEnergyConsumptionParams

	headers := ctx.Request().Header
	// ------------- Optional header parameter "x-correlator" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("x-correlator")]; found {
		var XCorrelator XCorrelator
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for x-correlator, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "x-correlator", valueList[0], &XCorrelator, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: false})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter x-correlator: %s", err))
		}

		params.XCorrelator = &XCorrelator
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.CalculateEnergyConsumption(ctx, params)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/calculate-carbon-footprint", wrapper.CalculateCarbonFootprint)
	router.POST(baseURL+"/calculate-energy-consumption", wrapper.CalculateEnergyConsumption)

}

type Generic400ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic400JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic400ResponseHeaders
}

type Generic401ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic401JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic401ResponseHeaders
}

type Generic403ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic403JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic403ResponseHeaders
}

type Generic404ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic404JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic404ResponseHeaders
}

type Generic410ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic410JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic410ResponseHeaders
}

type Generic429ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic429JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message A human-readable description of what the event represents
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic429ResponseHeaders
}

type CalculateCarbonFootprintRequestObject struct {
	Params CalculateCarbonFootprintParams
	Body   *CalculateCarbonFootprintJSONRequestBody
}

type CalculateCarbonFootprintResponseObject interface {
	VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error
}

type CalculateCarbonFootprint201ResponseHeaders struct {
	Location    string
	XCorrelator XCorrelator
}

type CalculateCarbonFootprint201JSONResponse struct {
	Body    ReportCreationRequest
	Headers CalculateCarbonFootprint201ResponseHeaders
}

func (response CalculateCarbonFootprint201JSONResponse) VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprint(response.Headers.Location))
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(201)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateCarbonFootprint400JSONResponse struct{ Generic400JSONResponse }

func (response CalculateCarbonFootprint400JSONResponse) VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateCarbonFootprint401JSONResponse struct{ Generic401JSONResponse }

func (response CalculateCarbonFootprint401JSONResponse) VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateCarbonFootprint403JSONResponse struct{ Generic403JSONResponse }

func (response CalculateCarbonFootprint403JSONResponse) VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateCarbonFootprint404JSONResponse struct{ Generic404JSONResponse }

func (response CalculateCarbonFootprint404JSONResponse) VisitCalculateCarbonFootprintResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateEnergyConsumptionRequestObject struct {
	Params CalculateEnergyConsumptionParams
	Body   *CalculateEnergyConsumptionJSONRequestBody
}

type CalculateEnergyConsumptionResponseObject interface {
	VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error
}

type CalculateEnergyConsumption201ResponseHeaders struct {
	Location    string
	XCorrelator XCorrelator
}

type CalculateEnergyConsumption201JSONResponse struct {
	Body    ReportCreationRequest
	Headers CalculateEnergyConsumption201ResponseHeaders
}

func (response CalculateEnergyConsumption201JSONResponse) VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprint(response.Headers.Location))
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(201)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateEnergyConsumption400JSONResponse struct{ Generic400JSONResponse }

func (response CalculateEnergyConsumption400JSONResponse) VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateEnergyConsumption401JSONResponse struct{ Generic401JSONResponse }

func (response CalculateEnergyConsumption401JSONResponse) VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateEnergyConsumption403JSONResponse struct{ Generic403JSONResponse }

func (response CalculateEnergyConsumption403JSONResponse) VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response.Body)
}

type CalculateEnergyConsumption404JSONResponse struct{ Generic404JSONResponse }

func (response CalculateEnergyConsumption404JSONResponse) VisitCalculateEnergyConsumptionResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Retrieves the overall carbon footprint for the target Application instances in a certain period of time.
	// (POST /calculate-carbon-footprint)
	CalculateCarbonFootprint(ctx context.Context, request CalculateCarbonFootprintRequestObject) (CalculateCarbonFootprintResponseObject, error)
	// Provides the overall Energy Consumption for the target Application instances in a certain period of time.
	// (POST /calculate-energy-consumption)
	CalculateEnergyConsumption(ctx context.Context, request CalculateEnergyConsumptionRequestObject) (CalculateEnergyConsumptionResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// CalculateCarbonFootprint operation middleware
func (sh *strictHandler) CalculateCarbonFootprint(ctx echo.Context, params CalculateCarbonFootprintParams) error {
	var request CalculateCarbonFootprintRequestObject

	request.Params = params

	var body CalculateCarbonFootprintJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.CalculateCarbonFootprint(ctx.Request().Context(), request.(CalculateCarbonFootprintRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "CalculateCarbonFootprint")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(CalculateCarbonFootprintResponseObject); ok {
		return validResponse.VisitCalculateCarbonFootprintResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// CalculateEnergyConsumption operation middleware
func (sh *strictHandler) CalculateEnergyConsumption(ctx echo.Context, params CalculateEnergyConsumptionParams) error {
	var request CalculateEnergyConsumptionRequestObject

	request.Params = params

	var body CalculateEnergyConsumptionJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.CalculateEnergyConsumption(ctx.Request().Context(), request.(CalculateEnergyConsumptionRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "CalculateEnergyConsumption")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(CalculateEnergyConsumptionResponseObject); ok {
		return validResponse.VisitCalculateEnergyConsumptionResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
