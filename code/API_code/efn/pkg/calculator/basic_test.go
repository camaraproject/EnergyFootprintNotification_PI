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
package calculator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
)

const tCO2ePerKWh = 0.00035

func floatPtr(f float64) *float64 { return &f }

func TestCalculateEnergyConsumption(t *testing.T) {
	tests := []struct {
		name      string
		input     []database.JobAppResult
		expects   *float64
		expectErr bool
	}{
		{
			name: "aggregates consumption across multiple apps with multiple NEs",
			input: []database.JobAppResult{
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app1",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: floatPtr(0.002),
						NetworkElements: map[string]database.NetworkElementResult{
							"ne1": {
								EnergyConsumption:  floatPtr(0.001),
								AppInstanceTraffic: floatPtr(100),
								TotalTraffic:       floatPtr(1000),
							},
							"ne2": {
								EnergyConsumption:  floatPtr(0.001),
								AppInstanceTraffic: floatPtr(100),
								TotalTraffic:       floatPtr(1000),
							},
						},
					},
				},
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app2",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: floatPtr(0.002),
						NetworkElements: map[string]database.NetworkElementResult{
							"ne1": {
								EnergyConsumption:  floatPtr(0.001),
								AppInstanceTraffic: floatPtr(100),
								TotalTraffic:       floatPtr(1000),
							},
							"ne2": {
								EnergyConsumption:  floatPtr(0.001),
								AppInstanceTraffic: floatPtr(100),
								TotalTraffic:       floatPtr(1000),
							},
						},
					},
				},
			},
			expects: floatPtr(0.0044),
		},
		{
			name: "all data present, single app, single NE",
			input: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					JobID: "job1",
					AppID: "app1",
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(2.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(100.0),
							TotalTraffic:       floatPtr(1000.0),
						},
					},
				},
			}},
			expects: floatPtr(2.1), // 2.0 + (1.0 * 0.1)
		},
		{
			name: "all data present, multiple apps and NEs",
			input: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					JobID: "job1",
					AppID: "app1",
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(2.0),
							AppInstanceTraffic: floatPtr(50.0),
							TotalTraffic:       floatPtr(100.0),
						},
						"ne2": {
							EnergyConsumption:  floatPtr(3.0),
							AppInstanceTraffic: floatPtr(10.0),
							TotalTraffic:       floatPtr(100.0),
						},
					},
				},
			}, {
				JobAppResultMetadata: database.JobAppResultMetadata{
					JobID: "job1",
					AppID: "app2",
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(2.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(20.0),
							TotalTraffic:       floatPtr(100.0),
						},
					},
				},
			}},
			expects: floatPtr(1.0 + (2.0 * 0.5) + (3.0 * 0.1) + 2.0 + (1.0 * 0.2)),
		},
		{
			name: "missing app instance energy consumption",
			input: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					JobID: "job1",
					AppID: "app1",
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: nil,
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(100.0),
							TotalTraffic:       floatPtr(1000.0),
						},
					},
				},
			}},
			expects:   nil,
			expectErr: true,
		},
		{
			name: "missing NE field",
			input: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					JobID: "job1",
					AppID: "app1",
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  nil,
							AppInstanceTraffic: floatPtr(100.0),
							TotalTraffic:       floatPtr(1000.0),
						},
					},
				},
			}},
			expects:   nil,
			expectErr: true,
		},
	}

	client := NewSimpleClient(tCO2ePerKWh)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CalculateEnergyConsumption(ctx, tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, *tt.expects, *result, 1e-6)
			}
		})
	}
}

func TestCalculateCarbonFootprint(t *testing.T) {
	tests := []struct {
		name      string
		input     []database.JobAppResult
		expects   *float64
		expectErr bool
	}{
		{
			name: "converts energy to carbon footprint correctly",
			input: []database.JobAppResult{
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app1",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: floatPtr(1.0), // 1 kWh
						NetworkElements: map[string]database.NetworkElementResult{
							"ne1": {
								EnergyConsumption:  floatPtr(1.0),
								AppInstanceTraffic: floatPtr(100),
								TotalTraffic:       floatPtr(1000),
							},
						},
					},
				},
			},
			// Energy: 1.0 + (1.0 * 0.1) = 1.1 kWh
			// Carbon: 1.1 * 0.00035 = 0.000385 tCO2e
			expects: floatPtr(1.1 * tCO2ePerKWh),
		},
		{
			name: "multiple apps carbon footprint",
			input: []database.JobAppResult{
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app1",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: floatPtr(2.0),
						NetworkElements: map[string]database.NetworkElementResult{
							"ne1": {
								EnergyConsumption:  floatPtr(1.0),
								AppInstanceTraffic: floatPtr(500),
								TotalTraffic:       floatPtr(1000),
							},
						},
					},
				},
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app2",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: floatPtr(3.0),
						NetworkElements: map[string]database.NetworkElementResult{
							"ne1": {
								EnergyConsumption:  floatPtr(2.0),
								AppInstanceTraffic: floatPtr(200),
								TotalTraffic:       floatPtr(1000),
							},
						},
					},
				},
			},
			// Energy: 2.0 + (1.0 * 0.5) + 3.0 + (2.0 * 0.2) = 5.9 kWh
			// Carbon: 5.9 * 0.00035 = 0.002065 tCO2e
			expects: floatPtr((2.0 + 0.5 + 3.0 + 0.4) * tCO2ePerKWh),
		},
		{
			name: "missing data returns error",
			input: []database.JobAppResult{
				{
					JobAppResultMetadata: database.JobAppResultMetadata{
						JobID: "job1",
						AppID: "app1",
					},
					Result: &database.TaskResult{
						AppInstanceEnergyConsumption: nil,
						NetworkElements:              map[string]database.NetworkElementResult{},
					},
				},
			},
			expects:   nil,
			expectErr: true,
		},
	}

	client := NewSimpleClient(tCO2ePerKWh)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CalculateCarbonFootprint(ctx, tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, *tt.expects, *result, 1e-6)
			}
		})
	}
}
