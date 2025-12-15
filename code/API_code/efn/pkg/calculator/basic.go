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
	"fmt"

	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
)

var _ Interface = &simpleClient{}

// simpleClient holds configuration values for calculations
type simpleClient struct {
	// CO2 conversion factor in tCO2e per kWh (e.g., 0.00035)
	tCO2ePerKWh float64
}

// NewSimpleClient creates a new calculator with the provided CO2 conversion factor.
// If factor <= 0, a default value of 0.00035 tCO2e/kWh is used.
func NewSimpleClient(factor float64) *simpleClient {
	if factor <= 0 {
		factor = 0.00035
	}
	return &simpleClient{tCO2ePerKWh: factor}
}

// CalculateEnergyConsumption calculates the energy consumption (kWh) for the given application instance, taking into account both application and all network elements consumption.
func (simpleClient) CalculateEnergyConsumption(ctx context.Context, data []database.JobAppResult) (*float64, error) {
	var result float64

	if err := validateInput(data[0]); err != nil {
		return nil, err
	}

	// Simple calculation: sum of app instance and proportional network element energy consumption
	for _, d := range data {
		appResult := d.Result
		result += *appResult.AppInstanceEnergyConsumption
		for _, ne := range appResult.NetworkElements {
			trafficShare := *ne.AppInstanceTraffic / *ne.TotalTraffic
			allocatedNEConsumption := *ne.EnergyConsumption * trafficShare
			result += allocatedNEConsumption
		}
	}
	return &result, nil
}

// CalculateCarbonFootprint calculates the carbon footprint (tCO2e) by first calculating
// energy consumption and then applying the CO2 conversion factor.
func (c simpleClient) CalculateCarbonFootprint(ctx context.Context, data []database.JobAppResult) (*float64, error) {
	energyKWh, err := c.CalculateEnergyConsumption(ctx, data)
	if err != nil {
		return nil, err
	}

	carbonFootprint := *energyKWh * c.tCO2ePerKWh
	return &carbonFootprint, nil
}

func validateInput(data database.JobAppResult) error {
	log := logger.Get().With(zap.String("requestID", data.JobID), zap.String("applicationInstanceID", data.AppID))
	if data.Result.AppInstanceEnergyConsumption == nil {
		msg := "Missing energy consumption for application instance"
		log.Error(msg)
		return fmt.Errorf("%s", msg)
	}

	for _, ne := range data.Result.NetworkElements {
		if ne.EnergyConsumption == nil || ne.TotalTraffic == nil || ne.AppInstanceTraffic == nil {
			msg := "Missing energy consumption for network element"
			log.With(zap.Any("networkElement", ne)).Error(msg)
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}
