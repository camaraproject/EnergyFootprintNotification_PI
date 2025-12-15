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

var _ Interface = &dummyClient{}

type dummyClient struct{}

func NewDummyClient() (*dummyClient, error) {
	return &dummyClient{}, nil
}

// RetrieveTrafficVolumes returns dummy traffic volume measurements for the given network elements.
// Returns both app instance IP volume and total NE volume for each network element.
// In a real implementation, appInstanceIPList would be used to filter traffic by the app's IPs.
func (k *dummyClient) RetrieveTrafficVolumes(ctx context.Context, appInstanceIPList []string, networkElements []NetworkElement, timePeriod *models.TimePeriod) (*TrafficVolumeMeasureList, error) {
	measures := make([]TrafficVolumeMeasure, 0, len(networkElements))

	for _, ne := range networkElements {
		measures = append(measures, TrafficVolumeMeasure{
			NetworkElement: NetworkElement{
				VendorIdentifier: ne.VendorIdentifier,
				NEIdentifier:     ne.NEIdentifier,
			},
			TrafficVolumeIP:  100.0,  // Dummy app instance IP volume in Mbps
			TrafficVolumeAll: 1000.0, // Dummy total NE volume in Mbps
		})
	}

	return &TrafficVolumeMeasureList{
		TrafficVolumeMeasureList: measures,
	}, nil
}
