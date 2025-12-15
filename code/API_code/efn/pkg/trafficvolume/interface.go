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

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

// NetworkElement identifies a specific network element by vendor and NE identifier.
type NetworkElement struct {
	VendorIdentifier string `json:"vendorIdentifier"`
	NEIdentifier     string `json:"neIdentifier"`
}

// TrafficVolumeMeasure contains volume measurements for a specific network element.
type TrafficVolumeMeasure struct {
	NetworkElement   NetworkElement `json:"networkElement"`
	TrafficVolumeIP  float64        `json:"trafficVolumeIP"`  // Volume for the app instance IP
	TrafficVolumeAll float64        `json:"trafficVolumeAll"` // Total volume for the network element
}

// TrafficVolumeMeasureList is the response containing measurements for multiple network elements.
type TrafficVolumeMeasureList struct {
	TrafficVolumeMeasureList []TrafficVolumeMeasure `json:"TrafficVolumeMeasureList"`
}

type Interface interface {
	// RetrieveTrafficVolumes retrieves traffic volume measurements for a list of network elements.
	// It returns both the app instance IP volume and total NE volume for each network element.
	// appInstanceIPList contains all IPs associated with the application instance.
	//
	// Expected request format for real implementation:
	// [
	//   {
	//     "appInstanceIdIpList": ["84.125.93.10", "84.125.93.11"],
	//     "networkElementList": [
	//       {"vendorIdentifier": "Ericsson", "neIdentifier": "pcg101"},
	//       {"vendorIdentifier": "Nokia", "neIdentifier": "pcg201"}
	//     ],
	//     "timePeriod": {
	//       "startDate": "2025-11-01T00:00:00Z",
	//       "endDate": "2025-11-02T00:00:00Z"
	//     }
	//   }
	// ]
	RetrieveTrafficVolumes(ctx context.Context, appInstanceIPList []string, networkElements []NetworkElement, timePeriod *models.TimePeriod) (*TrafficVolumeMeasureList, error)
}
