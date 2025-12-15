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
package orchestrator

import "context"

type dummyClient struct{}

func NewDummyClient() (*dummyClient, error) {
	return &dummyClient{}, nil
}

func (k *dummyClient) GatherInformation(_ context.Context, appID string) (Information, error) {
	// Stub implementation returning dummy data
	return Information{
		App: ApplicationInstance{
			IPList:    []string{"84.125.93.10", "84.125.93.11"},
			InfraType: "sylva",
		},
		NE: []NEInfo{
			{
				NetworkID:  appID + "-ne-1",
				VendorID:   "vendor-1",
				InstanceID: "ne-instance-1",
				InfraType:  "UPF-1",
			},
			{
				NetworkID:  appID + "-ne-2",
				VendorID:   "vendor-2",
				InstanceID: "ne-instance-2",
				InfraType:  "UPF-2",
			},
		},
	}, nil
}
