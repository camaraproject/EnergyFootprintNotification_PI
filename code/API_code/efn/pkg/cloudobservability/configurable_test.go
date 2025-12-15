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
	"testing"
	"time"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

func TestNewConfigurableClient(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		wantAppValue float64
		wantNEValue  float64
		wantSuccess  int
		wantError    int
		wantErrType  string
	}{
		{
			name:         "default values",
			envVars:      map[string]string{},
			wantAppValue: 0.0020,
			wantNEValue:  0.0010,
			wantSuccess:  0,
			wantError:    0,
			wantErrType:  "throttling",
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"CLOUDOBS_CONFIG_APP_VALUE":     "0.5",
				"CLOUDOBS_CONFIG_NE_VALUE":      "0.3",
				"CLOUDOBS_CONFIG_SUCCESS_COUNT": "10",
				"CLOUDOBS_CONFIG_ERROR_COUNT":   "5",
				"CLOUDOBS_CONFIG_ERROR_TYPE":    "throttling",
			},
			wantAppValue: 0.5,
			wantNEValue:  0.3,
			wantSuccess:  10,
			wantError:    5,
			wantErrType:  "throttling",
		},
		{
			name: "invalid values fallback to defaults",
			envVars: map[string]string{
				"CLOUDOBS_CONFIG_APP_VALUE":     "invalid",
				"CLOUDOBS_CONFIG_NE_VALUE":      "invalid",
				"CLOUDOBS_CONFIG_SUCCESS_COUNT": "invalid",
				"CLOUDOBS_CONFIG_ERROR_COUNT":   "invalid",
				"CLOUDOBS_CONFIG_ERROR_TYPE":    "invalid",
			},
			wantAppValue: 0.0020,
			wantNEValue:  0.0010,
			wantSuccess:  0,
			wantError:    0,
			wantErrType:  "throttling",
		},
		{
			name: "only errors configured",
			envVars: map[string]string{
				"CLOUDOBS_CONFIG_ERROR_COUNT": "3",
			},
			wantAppValue: 0.0020,
			wantNEValue:  0.0010,
			wantSuccess:  0,
			wantError:    3,
			wantErrType:  "throttling",
		},
		{
			name: "only success count configured",
			envVars: map[string]string{
				"CLOUDOBS_CONFIG_SUCCESS_COUNT": "5",
			},
			wantAppValue: 0.0020,
			wantNEValue:  0.0010,
			wantSuccess:  5,
			wantError:    0,
			wantErrType:  "throttling",
		},
		{
			name: "with delay configured",
			envVars: map[string]string{
				"CLOUDOBS_CONFIG_DELAY_MS": "100",
			},
			wantAppValue: 0.0020,
			wantNEValue:  0.0010,
			wantSuccess:  0,
			wantError:    0,
			wantErrType:  "throttling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, key := range []string{"CLOUDOBS_CONFIG_APP_VALUE", "CLOUDOBS_CONFIG_NE_VALUE",
				"CLOUDOBS_CONFIG_SUCCESS_COUNT", "CLOUDOBS_CONFIG_ERROR_COUNT", "CLOUDOBS_CONFIG_ERROR_TYPE"} {
				os.Unsetenv(key)
			}

			// Set test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			client, err := NewConfigurableClient()
			if err != nil {
				t.Fatalf("NewConfigurableClient() error = %v", err)
			}

			if client.appValue != tt.wantAppValue {
				t.Errorf("appValue = %v, want %v", client.appValue, tt.wantAppValue)
			}
			if client.neValue != tt.wantNEValue {
				t.Errorf("neValue = %v, want %v", client.neValue, tt.wantNEValue)
			}
			if client.successCount != tt.wantSuccess {
				t.Errorf("successCount = %v, want %v", client.successCount, tt.wantSuccess)
			}
			if client.errorCount != tt.wantError {
				t.Errorf("errorCount = %v, want %v", client.errorCount, tt.wantError)
			}
			if client.errorType != tt.wantErrType {
				t.Errorf("errorType = %v, want %v", client.errorType, tt.wantErrType)
			}
		})
	}
}

func TestConfigurableClient_RetrieveAppEnergyConsumption(t *testing.T) {
	tests := []struct {
		name         string
		successCount int
		errorCount   int
		errorType    string
		appValue     float64
		requests     int
		wantResults  []testResult
	}{
		{
			name:         "no errors configured",
			successCount: 0,
			errorCount:   0,
			appValue:     0.5,
			requests:     5,
			wantResults: []testResult{
				{wantValue: 0.5, wantErr: false},
				{wantValue: 0.5, wantErr: false},
				{wantValue: 0.5, wantErr: false},
				{wantValue: 0.5, wantErr: false},
				{wantValue: 0.5, wantErr: false},
			},
		},
		{
			name:         "always fail when successCount=0 and errorCount>0",
			successCount: 0,
			errorCount:   2,
			errorType:    "permanent",
			appValue:     0.5,
			requests:     4,
			wantResults: []testResult{
				{wantErr: true, wantErrType: "permanent"},
				{wantErr: true, wantErrType: "permanent"},
				{wantErr: true, wantErrType: "permanent"},
				{wantErr: true, wantErrType: "permanent"},
			},
		},
		{
			name:         "always succeed when successCount>0",
			successCount: 2,
			errorCount:   3,
			errorType:    "throttling",
			appValue:     0.3,
			requests:     7,
			wantResults: []testResult{
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
				{wantValue: 0.3, wantErr: false},
			},
		},
		{
			name:         "only success count no errors",
			successCount: 5,
			errorCount:   0,
			appValue:     0.2,
			requests:     7,
			wantResults: []testResult{
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
				{wantValue: 0.2, wantErr: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				appValue:     tt.appValue,
				neValue:      0.0010,
				successCount: tt.successCount,
				errorCount:   tt.errorCount,
				errorType:    tt.errorType,
				requestCount: 0,
			}

			ctx := context.Background()
			timePeriod := &models.TimePeriod{}

			for i, want := range tt.wantResults {
				result, err := client.RetrieveAppEnergyConsumption(ctx, "app-123", timePeriod, "vm")

				if want.wantErr {
					if err == nil {
						t.Errorf("request %d: expected error but got none", i+1)
					} else if want.wantErrType == "throttling" {
						if !IsThrottlingError(err) {
							t.Errorf("request %d: expected throttling error, got: %v", i+1, err)
						}
					} else {
						if IsThrottlingError(err) {
							t.Errorf("request %d: expected permanent error, got throttling error", i+1)
						}
					}
				} else {
					if err != nil {
						t.Errorf("request %d: unexpected error: %v", i+1, err)
					}
					if result == nil {
						t.Errorf("request %d: expected result but got nil", i+1)
					} else if *result != want.wantValue {
						t.Errorf("request %d: got value %v, want %v", i+1, *result, want.wantValue)
					}
				}
			}
		})
	}
}

func TestConfigurableClient_RetrieveNetworkElementEnergyConsumption(t *testing.T) {
	tests := []struct {
		name         string
		successCount int
		errorCount   int
		errorType    string
		neValue      float64
		requests     int
		wantResults  []testResult
	}{
		{
			name:         "no errors configured",
			successCount: 0,
			errorCount:   0,
			neValue:      0.8,
			requests:     3,
			wantResults: []testResult{
				{wantValue: 0.8, wantErr: false},
				{wantValue: 0.8, wantErr: false},
				{wantValue: 0.8, wantErr: false},
			},
		},
		{
			name:         "always fail when successCount=0 and errorCount>0",
			successCount: 0,
			errorCount:   1,
			errorType:    "permanent",
			neValue:      0.8,
			requests:     3,
			wantResults: []testResult{
				{wantErr: true, wantErrType: "permanent"},
				{wantErr: true, wantErrType: "permanent"},
				{wantErr: true, wantErrType: "permanent"},
			},
		},
		{
			name:         "always succeed when successCount>0",
			successCount: 3,
			errorCount:   2,
			errorType:    "throttling",
			neValue:      0.6,
			requests:     6,
			wantResults: []testResult{
				{wantValue: 0.6, wantErr: false},
				{wantValue: 0.6, wantErr: false},
				{wantValue: 0.6, wantErr: false},
				{wantValue: 0.6, wantErr: false},
				{wantValue: 0.6, wantErr: false},
				{wantValue: 0.6, wantErr: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				appValue:     0.0020,
				neValue:      tt.neValue,
				successCount: tt.successCount,
				errorCount:   tt.errorCount,
				errorType:    tt.errorType,
				requestCount: 0,
			}

			ctx := context.Background()
			timePeriod := &models.TimePeriod{}

			for i, want := range tt.wantResults {
				result, err := client.RetrieveNetworkElementEnergyConsumption(ctx, "app-123", timePeriod, "router")

				if want.wantErr {
					if err == nil {
						t.Errorf("request %d: expected error but got none", i+1)
					} else if want.wantErrType == "throttling" {
						if !IsThrottlingError(err) {
							t.Errorf("request %d: expected throttling error, got: %v", i+1, err)
						}
					} else {
						if IsThrottlingError(err) {
							t.Errorf("request %d: expected permanent error, got throttling error", i+1)
						}
					}
				} else {
					if err != nil {
						t.Errorf("request %d: unexpected error: %v", i+1, err)
					}
					if result == nil {
						t.Errorf("request %d: expected result but got nil", i+1)
					} else if *result != want.wantValue {
						t.Errorf("request %d: got value %v, want %v", i+1, *result, want.wantValue)
					}
				}
			}
		})
	}
}

func TestConfigurableClient_MakeError(t *testing.T) {
	tests := []struct {
		name          string
		errorType     string
		wantThrottle  bool
		requestNumber int
	}{
		{
			name:          "permanent error",
			errorType:     "permanent",
			wantThrottle:  false,
			requestNumber: 1,
		},
		{
			name:          "throttling error",
			errorType:     "throttling",
			wantThrottle:  true,
			requestNumber: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				errorType:    tt.errorType,
				requestCount: tt.requestNumber,
			}

			err := client.makeError()

			if err == nil {
				t.Fatal("expected error but got nil")
			}

			if IsThrottlingError(err) != tt.wantThrottle {
				t.Errorf("IsThrottlingError() = %v, want %v", IsThrottlingError(err), tt.wantThrottle)
			}

			expectedMsg := fmt.Sprintf("configurable %s error (request #%d)",
				map[bool]string{true: "throttling", false: "permanent"}[tt.wantThrottle],
				tt.requestNumber)

			if err.Error() != expectedMsg {
				t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
			}
		})
	}
}

func TestConfigurableClient_Delay(t *testing.T) {
	tests := []struct {
		name      string
		delayMS   int
		wantDelay time.Duration
	}{
		{
			name:      "no delay",
			delayMS:   0,
			wantDelay: 0,
		},
		{
			name:      "100ms delay",
			delayMS:   100,
			wantDelay: 100 * time.Millisecond,
		},
		{
			name:      "1 second delay",
			delayMS:   1000,
			wantDelay: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				appValue: 0.5,
				delay:    time.Duration(tt.delayMS) * time.Millisecond,
			}

			ctx := context.Background()
			timePeriod := &models.TimePeriod{}

			start := time.Now()
			_, err := client.RetrieveAppEnergyConsumption(ctx, "app-123", timePeriod, "vm")
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if elapsed < tt.wantDelay {
				t.Errorf("expected delay of at least %v, got %v", tt.wantDelay, elapsed)
			}
		})
	}
}

func TestConfigurableClient_DelayContextCancellation(t *testing.T) {
	client := &configurableClient{
		appValue: 0.5,
		delay:    500 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	timePeriod := &models.TimePeriod{}

	_, err := client.RetrieveAppEnergyConsumption(ctx, "app-123", timePeriod, "vm")

	if err == nil {
		t.Error("expected context cancellation error but got none")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

type testResult struct {
	wantValue   float64
	wantErr     bool
	wantErrType string
}
