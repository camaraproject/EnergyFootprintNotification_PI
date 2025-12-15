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
	"os"
	"testing"
	"time"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

func TestNewConfigurableClient(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		wantIPVolume  float64
		wantAllVolume float64
		wantSuccess   int
		wantError     int
		wantThrottle  bool
	}{
		{
			name:          "default values (throttling enabled)",
			envVars:       map[string]string{},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  true,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_IP_VOLUME":     "250.5",
				"TRAFFIC_CONFIG_ALL_VOLUME":    "5000.0",
				"TRAFFIC_CONFIG_SUCCESS_COUNT": "8",
				"TRAFFIC_CONFIG_ERROR_COUNT":   "4",
				"TRAFFIC_CONFIG_ERROR_TYPE":    "permanent",
			},
			wantIPVolume:  250.5,
			wantAllVolume: 5000.0,
			wantSuccess:   8,
			wantError:     4,
			wantThrottle:  false,
		},
		{
			name: "invalid values fallback to defaults (throttling enabled)",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_IP_VOLUME":     "invalid",
				"TRAFFIC_CONFIG_ALL_VOLUME":    "invalid",
				"TRAFFIC_CONFIG_SUCCESS_COUNT": "invalid",
				"TRAFFIC_CONFIG_ERROR_COUNT":   "invalid",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  true,
		},
		{
			name: "only errors configured (throttling enabled)",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_ERROR_COUNT": "5",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     5,
			wantThrottle:  true,
		},
		{
			name: "only success count configured (throttling enabled)",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_SUCCESS_COUNT": "10",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   10,
			wantError:     0,
			wantThrottle:  true,
		},
		{
			name: "zero values explicit (throttling enabled)",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_IP_VOLUME":  "0",
				"TRAFFIC_CONFIG_ALL_VOLUME": "0",
			},
			wantIPVolume:  0.0,
			wantAllVolume: 0.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  true,
		},
		{
			name: "with delay configured (throttling enabled)",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_DELAY_MS": "50",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  true,
		},
		{
			name: "explicitly disable throttling and enable permanent errors",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_ERROR_TYPE": "permanent",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  false,
		},
		{
			name: "explicitly enable throttling",
			envVars: map[string]string{
				"TRAFFIC_CONFIG_ERROR_TYPE": "throttling",
			},
			wantIPVolume:  100.0,
			wantAllVolume: 1000.0,
			wantSuccess:   0,
			wantError:     0,
			wantThrottle:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{
				"TRAFFIC_CONFIG_IP_VOLUME",
				"TRAFFIC_CONFIG_ALL_VOLUME",
				"TRAFFIC_CONFIG_SUCCESS_COUNT",
				"TRAFFIC_CONFIG_ERROR_COUNT",
				"TRAFFIC_CONFIG_DELAY_MS",
				"TRAFFIC_CONFIG_ERROR_TYPE",
			} {
				os.Unsetenv(key)
			}

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

			if client.ipVolume != tt.wantIPVolume {
				t.Errorf("ipVolume = %v, want %v", client.ipVolume, tt.wantIPVolume)
			}
			if client.allVolume != tt.wantAllVolume {
				t.Errorf("allVolume = %v, want %v", client.allVolume, tt.wantAllVolume)
			}
			if client.successCount != tt.wantSuccess {
				t.Errorf("successCount = %v, want %v", client.successCount, tt.wantSuccess)
			}
			if client.errorCount != tt.wantError {
				t.Errorf("errorCount = %v, want %v", client.errorCount, tt.wantError)
			}
			if (client.errorType == "throttling") != tt.wantThrottle {
				t.Errorf("throttle = %v, want %v", (client.errorType == "throttling"), tt.wantThrottle)
			}
		})
	}
}

func TestConfigurableClient_RetrieveTrafficVolumes(t *testing.T) {
	tests := []struct {
		name            string
		successCount    int
		errorCount      int
		ipVolume        float64
		allVolume       float64
		networkElements []NetworkElement
		requests        int
		wantResults     []trafficTestResult
	}{
		{
			name:         "no errors configured",
			successCount: 0,
			errorCount:   0,
			ipVolume:     150.0,
			allVolume:    1500.0,
			networkElements: []NetworkElement{
				{VendorIdentifier: "Ericsson", NEIdentifier: "ne1"},
				{VendorIdentifier: "Nokia", NEIdentifier: "ne2"},
			},
			requests: 3,
			wantResults: []trafficTestResult{
				{wantErr: false, wantIPVolume: 150.0, wantAllVolume: 1500.0, wantCount: 2},
				{wantErr: false, wantIPVolume: 150.0, wantAllVolume: 1500.0, wantCount: 2},
				{wantErr: false, wantIPVolume: 150.0, wantAllVolume: 1500.0, wantCount: 2},
			},
		},
		{
			name:         "always fail when successCount=0 and errorCount>0",
			successCount: 0,
			errorCount:   2,
			ipVolume:     200.0,
			allVolume:    2000.0,
			networkElements: []NetworkElement{
				{VendorIdentifier: "Vendor1", NEIdentifier: "ne1"},
			},
			requests: 4,
			wantResults: []trafficTestResult{
				{wantErr: true},
				{wantErr: true},
				{wantErr: true},
				{wantErr: true},
			},
		},
		{
			name:         "always succeed when successCount>0",
			successCount: 3,
			errorCount:   2,
			ipVolume:     75.5,
			allVolume:    755.0,
			networkElements: []NetworkElement{
				{VendorIdentifier: "Ericsson", NEIdentifier: "pcg101"},
				{VendorIdentifier: "Nokia", NEIdentifier: "pcg201"},
				{VendorIdentifier: "Huawei", NEIdentifier: "pcg301"},
			},
			requests: 7,
			wantResults: []trafficTestResult{
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
				{wantErr: false, wantIPVolume: 75.5, wantAllVolume: 755.0, wantCount: 3},
			},
		},
		{
			name:         "only success count no errors",
			successCount: 5,
			errorCount:   0,
			ipVolume:     100.0,
			allVolume:    1000.0,
			networkElements: []NetworkElement{
				{VendorIdentifier: "Test", NEIdentifier: "test1"},
			},
			requests: 7,
			wantResults: []trafficTestResult{
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 1},
			},
		},
		{
			name:            "empty network elements",
			successCount:    0,
			errorCount:      0,
			ipVolume:        100.0,
			allVolume:       1000.0,
			networkElements: []NetworkElement{},
			requests:        2,
			wantResults: []trafficTestResult{
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 0},
				{wantErr: false, wantIPVolume: 100.0, wantAllVolume: 1000.0, wantCount: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				ipVolume:     tt.ipVolume,
				allVolume:    tt.allVolume,
				successCount: tt.successCount,
				errorCount:   tt.errorCount,
				requestCount: 0,
			}

			ctx := context.Background()
			timePeriod := &models.TimePeriod{}
			appIPList := []string{"10.0.0.1", "10.0.0.2"}

			for i, want := range tt.wantResults {
				result, err := client.RetrieveTrafficVolumes(ctx, appIPList, tt.networkElements, timePeriod)

				if want.wantErr {
					if err == nil {
						t.Errorf("request %d: expected error but got none", i+1)
					}
				} else {
					if err != nil {
						t.Errorf("request %d: unexpected error: %v", i+1, err)
					}
					if result == nil {
						t.Errorf("request %d: expected result but got nil", i+1)
					} else {
						if len(result.TrafficVolumeMeasureList) != want.wantCount {
							t.Errorf("request %d: got %d measures, want %d", i+1, len(result.TrafficVolumeMeasureList), want.wantCount)
						}
						// Validate each measure has correct values
						for j, measure := range result.TrafficVolumeMeasureList {
							if measure.TrafficVolumeIP != want.wantIPVolume {
								t.Errorf("request %d, measure %d: TrafficVolumeIP = %v, want %v", i+1, j, measure.TrafficVolumeIP, want.wantIPVolume)
							}
							if measure.TrafficVolumeAll != want.wantAllVolume {
								t.Errorf("request %d, measure %d: TrafficVolumeAll = %v, want %v", i+1, j, measure.TrafficVolumeAll, want.wantAllVolume)
							}
							// Verify network element matches input
							if j < len(tt.networkElements) {
								if measure.NetworkElement.VendorIdentifier != tt.networkElements[j].VendorIdentifier {
									t.Errorf("request %d, measure %d: VendorIdentifier = %v, want %v", i+1, j, measure.NetworkElement.VendorIdentifier, tt.networkElements[j].VendorIdentifier)
								}
								if measure.NetworkElement.NEIdentifier != tt.networkElements[j].NEIdentifier {
									t.Errorf("request %d, measure %d: NEIdentifier = %v, want %v", i+1, j, measure.NetworkElement.NEIdentifier, tt.networkElements[j].NEIdentifier)
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestConfigurableClient_ShouldReturnError(t *testing.T) {
	tests := []struct {
		name         string
		successCount int
		errorCount   int
		requests     int
		wantErrors   []bool
	}{
		{
			name:         "no errors",
			successCount: 0,
			errorCount:   0,
			requests:     5,
			wantErrors:   []bool{false, false, false, false, false},
		},
		{
			name:         "all errors when successCount=0 and errorCount>0",
			successCount: 0,
			errorCount:   5,
			requests:     5,
			wantErrors:   []bool{true, true, true, true, true},
		},
		{
			name:         "always succeed when successCount>0",
			successCount: 2,
			errorCount:   2,
			requests:     6,
			wantErrors:   []bool{false, false, false, false, false, false},
		},
		{
			name:         "always succeed when successCount>0 and errorCount>0",
			successCount: 3,
			errorCount:   1,
			requests:     5,
			wantErrors:   []bool{false, false, false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				successCount: tt.successCount,
				errorCount:   tt.errorCount,
				requestCount: 0,
			}

			for i, wantErr := range tt.wantErrors {
				gotErr, err := client.shouldReturnError()
				if gotErr != wantErr {
					t.Errorf("request %d: shouldReturnError() = %v, want %v", i+1, gotErr, wantErr)
				}
				if gotErr && err == nil {
					t.Errorf("request %d: expected error object but got nil", i+1)
				}
				if !gotErr && err != nil {
					t.Errorf("request %d: unexpected error object: %v", i+1, err)
				}
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
			name:      "50ms delay",
			delayMS:   50,
			wantDelay: 50 * time.Millisecond,
		},
		{
			name:      "500ms delay",
			delayMS:   500,
			wantDelay: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &configurableClient{
				ipVolume:  100.0,
				allVolume: 1000.0,
				delay:     time.Duration(tt.delayMS) * time.Millisecond,
			}

			ctx := context.Background()
			timePeriod := &models.TimePeriod{}
			networkElements := []NetworkElement{
				{VendorIdentifier: "Test", NEIdentifier: "ne1"},
			}

			start := time.Now()
			_, err := client.RetrieveTrafficVolumes(ctx, []string{"10.0.0.1"}, networkElements, timePeriod)
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
		ipVolume:  100.0,
		allVolume: 1000.0,
		delay:     300 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	timePeriod := &models.TimePeriod{}
	networkElements := []NetworkElement{
		{VendorIdentifier: "Test", NEIdentifier: "ne1"},
	}

	_, err := client.RetrieveTrafficVolumes(ctx, []string{"10.0.0.1"}, networkElements, timePeriod)

	if err == nil {
		t.Error("expected context cancellation error but got none")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

type trafficTestResult struct {
	wantErr       bool
	wantIPVolume  float64
	wantAllVolume float64
	wantCount     int
}
