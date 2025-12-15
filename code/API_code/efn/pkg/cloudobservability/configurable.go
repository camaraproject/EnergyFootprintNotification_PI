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
package cloudobservability

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

var _ Interface = &configurableClient{}

// configurableClient allows configuring behavior via environment variables:
// - CLOUDOBS_CONFIG_APP_VALUE: Default app energy consumption value (default: 0.0020)
// - CLOUDOBS_CONFIG_NE_VALUE: Default NE energy consumption value (default: 0.0010)
// - CLOUDOBS_CONFIG_SUCCESS_COUNT: If >0, always succeed (default: 0)
// - CLOUDOBS_CONFIG_ERROR_COUNT: If >0 and successCount=0, always fail (default: 0)
// - CLOUDOBS_CONFIG_ERROR_TYPE: Type of error to return: "permanent" or "throttling" (default: "throttling")
// - CLOUDOBS_CONFIG_DELAY_MS: Request processing delay in milliseconds (default: 0)
//
// Behavior:
// - successCount=0, errorCount>0: always fail
// - successCount>0, errorCount=0: always succeed
// - successCount=0, errorCount=0: always succeed
type configurableClient struct {
	appValue     float64
	neValue      float64
	successCount int
	errorCount   int
	errorType    string
	delay        time.Duration
	requestCount int
	mu           sync.Mutex
}

func NewConfigurableClient() (*configurableClient, error) {
	appValue := 0.0020
	if val := os.Getenv("CLOUDOBS_CONFIG_APP_VALUE"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			appValue = parsed
		}
	}

	neValue := 0.0010
	if val := os.Getenv("CLOUDOBS_CONFIG_NE_VALUE"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			neValue = parsed
		}
	}

	successCount := 0
	if val := os.Getenv("CLOUDOBS_CONFIG_SUCCESS_COUNT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			successCount = parsed
		}
	}

	errorCount := 0
	if val := os.Getenv("CLOUDOBS_CONFIG_ERROR_COUNT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			errorCount = parsed
		}
	}

	errorType := "throttling"
	if val := os.Getenv("CLOUDOBS_CONFIG_ERROR_TYPE"); val != "" {
		if val == "throttling" || val == "permanent" {
			errorType = val
		}
	}

	delayMS := 0
	if val := os.Getenv("CLOUDOBS_CONFIG_DELAY_MS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			delayMS = parsed
		}
	}

	return &configurableClient{
		appValue:     appValue,
		neValue:      neValue,
		successCount: successCount,
		errorCount:   errorCount,
		errorType:    errorType,
		delay:        time.Duration(delayMS) * time.Millisecond,
		requestCount: 0,
	}, nil
}

func (c *configurableClient) shouldReturnError() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestCount++

	// successCount=0 and errorCount>0: always fail
	// successCount>0 or errorCount=0: always succeed
	if c.successCount == 0 && c.errorCount > 0 {
		return true, c.makeError()
	}

	return false, nil
}

func (c *configurableClient) makeError() error {
	if c.errorType == "throttling" {
		return NewThrottlingError(fmt.Sprintf("configurable throttling error (request #%d)", c.requestCount))
	}
	return fmt.Errorf("configurable permanent error (request #%d)", c.requestCount)
}

func (c *configurableClient) RetrieveAppEnergyConsumption(ctx context.Context, appInstanceID string, timePeriod *models.TimePeriod, appInfraType string) (*float64, error) {
	if c.delay > 0 {
		select {
		case <-time.After(c.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if shouldError, err := c.shouldReturnError(); shouldError {
		return nil, err
	}
	value := c.appValue
	return &value, nil
}

func (c *configurableClient) RetrieveNetworkElementEnergyConsumption(ctx context.Context, appInstanceID string, timePeriod *models.TimePeriod, neInfraType string) (*float64, error) {
	if c.delay > 0 {
		select {
		case <-time.After(c.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if shouldError, err := c.shouldReturnError(); shouldError {
		return nil, err
	}
	value := c.neValue
	return &value, nil
}
