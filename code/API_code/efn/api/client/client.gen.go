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
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"

	. "github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// CalculateCarbonFootprintWithBody request with any body
	CalculateCarbonFootprintWithBody(ctx context.Context, params *CalculateCarbonFootprintParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CalculateCarbonFootprint(ctx context.Context, params *CalculateCarbonFootprintParams, body CalculateCarbonFootprintJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CalculateEnergyConsumptionWithBody request with any body
	CalculateEnergyConsumptionWithBody(ctx context.Context, params *CalculateEnergyConsumptionParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CalculateEnergyConsumption(ctx context.Context, params *CalculateEnergyConsumptionParams, body CalculateEnergyConsumptionJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) CalculateCarbonFootprintWithBody(ctx context.Context, params *CalculateCarbonFootprintParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCalculateCarbonFootprintRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CalculateCarbonFootprint(ctx context.Context, params *CalculateCarbonFootprintParams, body CalculateCarbonFootprintJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCalculateCarbonFootprintRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CalculateEnergyConsumptionWithBody(ctx context.Context, params *CalculateEnergyConsumptionParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCalculateEnergyConsumptionRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CalculateEnergyConsumption(ctx context.Context, params *CalculateEnergyConsumptionParams, body CalculateEnergyConsumptionJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCalculateEnergyConsumptionRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewCalculateCarbonFootprintRequest calls the generic CalculateCarbonFootprint builder with application/json body
func NewCalculateCarbonFootprintRequest(server string, params *CalculateCarbonFootprintParams, body CalculateCarbonFootprintJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCalculateCarbonFootprintRequestWithBody(server, params, "application/json", bodyReader)
}

// NewCalculateCarbonFootprintRequestWithBody generates requests for CalculateCarbonFootprint with any type of body
func NewCalculateCarbonFootprintRequestWithBody(server string, params *CalculateCarbonFootprintParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/calculate-carbon-footprint")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	if params != nil {

		if params.XCorrelator != nil {
			var headerParam0 string

			headerParam0, err = runtime.StyleParamWithLocation("simple", false, "x-correlator", runtime.ParamLocationHeader, *params.XCorrelator)
			if err != nil {
				return nil, err
			}

			req.Header.Set("x-correlator", headerParam0)
		}

	}

	return req, nil
}

// NewCalculateEnergyConsumptionRequest calls the generic CalculateEnergyConsumption builder with application/json body
func NewCalculateEnergyConsumptionRequest(server string, params *CalculateEnergyConsumptionParams, body CalculateEnergyConsumptionJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCalculateEnergyConsumptionRequestWithBody(server, params, "application/json", bodyReader)
}

// NewCalculateEnergyConsumptionRequestWithBody generates requests for CalculateEnergyConsumption with any type of body
func NewCalculateEnergyConsumptionRequestWithBody(server string, params *CalculateEnergyConsumptionParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/calculate-energy-consumption")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	if params != nil {

		if params.XCorrelator != nil {
			var headerParam0 string

			headerParam0, err = runtime.StyleParamWithLocation("simple", false, "x-correlator", runtime.ParamLocationHeader, *params.XCorrelator)
			if err != nil {
				return nil, err
			}

			req.Header.Set("x-correlator", headerParam0)
		}

	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// CalculateCarbonFootprintWithBodyWithResponse request with any body
	CalculateCarbonFootprintWithBodyWithResponse(ctx context.Context, params *CalculateCarbonFootprintParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CalculateCarbonFootprintResponse, error)

	CalculateCarbonFootprintWithResponse(ctx context.Context, params *CalculateCarbonFootprintParams, body CalculateCarbonFootprintJSONRequestBody, reqEditors ...RequestEditorFn) (*CalculateCarbonFootprintResponse, error)

	// CalculateEnergyConsumptionWithBodyWithResponse request with any body
	CalculateEnergyConsumptionWithBodyWithResponse(ctx context.Context, params *CalculateEnergyConsumptionParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CalculateEnergyConsumptionResponse, error)

	CalculateEnergyConsumptionWithResponse(ctx context.Context, params *CalculateEnergyConsumptionParams, body CalculateEnergyConsumptionJSONRequestBody, reqEditors ...RequestEditorFn) (*CalculateEnergyConsumptionResponse, error)
}

type CalculateCarbonFootprintResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *ReportCreationRequest
	JSON400      *Generic400
	JSON401      *Generic401
	JSON403      *Generic403
	JSON404      *Generic404
}

// Status returns HTTPResponse.Status
func (r CalculateCarbonFootprintResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CalculateCarbonFootprintResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CalculateEnergyConsumptionResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *ReportCreationRequest
	JSON400      *Generic400
	JSON401      *Generic401
	JSON403      *Generic403
	JSON404      *Generic404
}

// Status returns HTTPResponse.Status
func (r CalculateEnergyConsumptionResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CalculateEnergyConsumptionResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// CalculateCarbonFootprintWithBodyWithResponse request with arbitrary body returning *CalculateCarbonFootprintResponse
func (c *ClientWithResponses) CalculateCarbonFootprintWithBodyWithResponse(ctx context.Context, params *CalculateCarbonFootprintParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CalculateCarbonFootprintResponse, error) {
	rsp, err := c.CalculateCarbonFootprintWithBody(ctx, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCalculateCarbonFootprintResponse(rsp)
}

func (c *ClientWithResponses) CalculateCarbonFootprintWithResponse(ctx context.Context, params *CalculateCarbonFootprintParams, body CalculateCarbonFootprintJSONRequestBody, reqEditors ...RequestEditorFn) (*CalculateCarbonFootprintResponse, error) {
	rsp, err := c.CalculateCarbonFootprint(ctx, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCalculateCarbonFootprintResponse(rsp)
}

// CalculateEnergyConsumptionWithBodyWithResponse request with arbitrary body returning *CalculateEnergyConsumptionResponse
func (c *ClientWithResponses) CalculateEnergyConsumptionWithBodyWithResponse(ctx context.Context, params *CalculateEnergyConsumptionParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CalculateEnergyConsumptionResponse, error) {
	rsp, err := c.CalculateEnergyConsumptionWithBody(ctx, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCalculateEnergyConsumptionResponse(rsp)
}

func (c *ClientWithResponses) CalculateEnergyConsumptionWithResponse(ctx context.Context, params *CalculateEnergyConsumptionParams, body CalculateEnergyConsumptionJSONRequestBody, reqEditors ...RequestEditorFn) (*CalculateEnergyConsumptionResponse, error) {
	rsp, err := c.CalculateEnergyConsumption(ctx, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCalculateEnergyConsumptionResponse(rsp)
}

// ParseCalculateCarbonFootprintResponse parses an HTTP response from a CalculateCarbonFootprintWithResponse call
func ParseCalculateCarbonFootprintResponse(rsp *http.Response) (*CalculateCarbonFootprintResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CalculateCarbonFootprintResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest ReportCreationRequest
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Generic400
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Generic401
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Generic403
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Generic404
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}

// ParseCalculateEnergyConsumptionResponse parses an HTTP response from a CalculateEnergyConsumptionWithResponse call
func ParseCalculateEnergyConsumptionResponse(rsp *http.Response) (*CalculateEnergyConsumptionResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CalculateEnergyConsumptionResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest ReportCreationRequest
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Generic400
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Generic401
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Generic403
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Generic404
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}
