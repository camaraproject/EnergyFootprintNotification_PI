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
package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
)

type mockDatabase struct {
	mock.Mock
	database.Interface
}

func (m *mockDatabase) GetAllJobAppResults(ctx context.Context, requestID string) ([]database.JobAppResult, error) {
	args := m.Called(ctx, requestID)
	return args.Get(0).([]database.JobAppResult), args.Error(1)
}

func (m *mockDatabase) GetJob(ctx context.Context, id string) (*database.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*database.Job), args.Error(1)
}

func (m *mockDatabase) TrySetCalculationTriggered(ctx context.Context, jobID string) (bool, error) {
	args := m.Called(ctx, jobID)
	return args.Bool(0), args.Error(1)
}

func (m *mockDatabase) TrySetNotificationSent(ctx context.Context, jobID string) (bool, error) {
	args := m.Called(ctx, jobID)
	return args.Bool(0), args.Error(1)
}

func TestIsAllDataGathered(t *testing.T) {
	app1 := uuid.New()
	app2 := uuid.New()
	tests := []struct {
		name       string
		results    []database.JobAppResult
		resultsErr error
		gottenJob  *database.Job
		modify     func(*database.TaskResult)
		expect     bool
		expectErr  bool
	}{
		{
			name: "all data gathered",
			gottenJob: &database.Job{
				JobSpec: database.JobSpec{
					Service: []models.AppInstanceId{app1},
				},
			},
			results: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					NumberOfTotalNEs: 1,
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(1.0),
							TotalTraffic:       floatPtr(1.0),
						},
					},
				},
			}},
			expect: true,
		},
		{
			name: "missing JobAppResult for an app",
			gottenJob: &database.Job{
				JobSpec: database.JobSpec{
					Service: []models.AppInstanceId{app1, app2},
				},
			},
			results: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					NumberOfTotalNEs: 1,
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(1.0),
							TotalTraffic:       floatPtr(1.0),
						},
					},
				},
			}},
			expect: false,
		},
		{
			name: "missing app instance energy consumption",
			gottenJob: &database.Job{
				JobSpec: database.JobSpec{
					Service: []models.AppInstanceId{app1},
				},
			},
			results: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					NumberOfTotalNEs: 1,
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: nil,
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(1.0),
							TotalTraffic:       floatPtr(1.0),
						},
					},
				},
			}},
			expect: false,
		},
		{
			name: "missing NE data",
			gottenJob: &database.Job{
				JobSpec: database.JobSpec{
					Service: []models.AppInstanceId{app1},
				},
			},
			results: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					NumberOfTotalNEs: 2,
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  floatPtr(1.0),
							AppInstanceTraffic: floatPtr(1.0),
							TotalTraffic:       floatPtr(1.0),
						},
					},
				},
			}},
			expect: false,
		},
		{
			name: "missing NE field",
			gottenJob: &database.Job{
				JobSpec: database.JobSpec{
					Service: []models.AppInstanceId{app1},
				},
			},
			results: []database.JobAppResult{{
				JobAppResultMetadata: database.JobAppResultMetadata{
					NumberOfTotalNEs: 1,
				},
				Result: &database.TaskResult{
					AppInstanceEnergyConsumption: floatPtr(1.0),
					NetworkElements: map[string]database.NetworkElementResult{
						"ne1": {
							EnergyConsumption:  nil,
							AppInstanceTraffic: floatPtr(1.0),
							TotalTraffic:       floatPtr(1.0),
						},
					},
				},
			}},
			expect: false,
		},
		{
			name:       "db error",
			resultsErr: errors.New("db error"),
			expect:     false,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDatabase{}
			db.On("GetAllJobAppResults", mock.Anything, mock.Anything).Return(tt.results, tt.resultsErr)
			db.On("GetJob", mock.Anything, mock.Anything).Return(tt.gottenJob, nil)
			// For these tests we want to allow triggering if data gathered.
			db.On("TrySetCalculationTriggered", mock.Anything, mock.Anything).Return(true, nil)
			h := &Handler{database: db}
			ok, err := h.isAllDataGathered(context.Background(), "req1")
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, ok)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
