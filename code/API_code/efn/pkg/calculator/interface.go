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

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
)

type Interface interface {
	// CalculateEnergyConsumption calculates the energy consumption (kWh) for the given application instance, taking into account both application and all network elements consumption.
	CalculateEnergyConsumption(ctx context.Context, data []database.JobAppResult) (*float64, error)

	// CalculateCarbonFootprint calculates the carbon footprint (tCO2e - tonnes of CO2 equivalent) for the given application instance.
	// It first calculates energy consumption, then converts it to CO2 equivalent using the conversion factor.
	CalculateCarbonFootprint(ctx context.Context, data []database.JobAppResult) (*float64, error)
}
