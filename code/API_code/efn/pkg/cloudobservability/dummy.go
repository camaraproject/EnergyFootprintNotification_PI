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

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

var _ Interface = &dummyClient{}

type dummyClient struct{}

func NewDummyClient() (*dummyClient, error) {
	return &dummyClient{}, nil
}

func (k *dummyClient) RetrieveAppEnergyConsumption(ctx context.Context, appInstanceID string, timePeriod *models.TimePeriod, appInfraType string) (*float64, error) {
	value := float64(0.0020)
	return &value, nil
}

func (k *dummyClient) RetrieveNetworkElementEnergyConsumption(ctx context.Context, appInstanceID string, timePeriod *models.TimePeriod, neInfraType string) (*float64, error) {
	value := float64(0.0010)
	return &value, nil
}
