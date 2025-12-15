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
package trafficvolume

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
// - TRAFFIC_CONFIG_IP_VOLUME: Default app instance IP traffic volume in Mbps (default: 100.0)
// - TRAFFIC_CONFIG_ALL_VOLUME: Default total NE traffic volume in Mbps (default: 1000.0)
// - TRAFFIC_CONFIG_SUCCESS_COUNT: If >0, always succeed (default: 0)
// - TRAFFIC_CONFIG_ERROR_COUNT: If >0 and successCount=0, always fail (default: 0)
// - TRAFFIC_CONFIG_ERROR_TYPE: Type of error to return: "permanent" or "throttling" (default: "throttling")
// - TRAFFIC_CONFIG_DELAY_MS: Request processing delay in milliseconds (default: 0)
//
// Behavior:
// - successCount=0, errorCount>0: always fail
// - successCount>0, errorCount=0: always succeed
// - successCount=0, errorCount=0: always succeed
type configurableClient struct {
	ipVolume     float64
	allVolume    float64
	successCount int
	errorCount   int
	errorType    string
	delay        time.Duration
	requestCount int
	mu           sync.Mutex
}

func NewConfigurableClient() (*configurableClient, error) {
	ipVolume := 100.0
	if val := os.Getenv("TRAFFIC_CONFIG_IP_VOLUME"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			ipVolume = parsed
		}
	}

	allVolume := 1000.0
	if val := os.Getenv("TRAFFIC_CONFIG_ALL_VOLUME"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			allVolume = parsed
		}
	}

	successCount := 0
	if val := os.Getenv("TRAFFIC_CONFIG_SUCCESS_COUNT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			successCount = parsed
		}
	}

	errorCount := 0
	if val := os.Getenv("TRAFFIC_CONFIG_ERROR_COUNT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			errorCount = parsed
		}
	}

	errorType := "throttling"
	if val := os.Getenv("TRAFFIC_CONFIG_ERROR_TYPE"); val != "" {
		if val == "throttling" || val == "permanent" {
			errorType = val
		}
	}

	delayMS := 0
	if val := os.Getenv("TRAFFIC_CONFIG_DELAY_MS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			delayMS = parsed
		}
	}

	return &configurableClient{
		ipVolume:     ipVolume,
		allVolume:    allVolume,
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

func (c *configurableClient) RetrieveTrafficVolumes(ctx context.Context, appInstanceIPList []string, networkElements []NetworkElement, timePeriod *models.TimePeriod) (*TrafficVolumeMeasureList, error) {
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

	measures := make([]TrafficVolumeMeasure, 0, len(networkElements))

	for _, ne := range networkElements {
		measures = append(measures, TrafficVolumeMeasure{
			NetworkElement: NetworkElement{
				VendorIdentifier: ne.VendorIdentifier,
				NEIdentifier:     ne.NEIdentifier,
			},
			TrafficVolumeIP:  c.ipVolume,
			TrafficVolumeAll: c.allVolume,
		})
	}

	return &TrafficVolumeMeasureList{
		TrafficVolumeMeasureList: measures,
	}, nil
}
