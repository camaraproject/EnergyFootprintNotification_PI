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

// Information is the consolidated view for a given AppInstanceID.
// It keeps app/runtime details separate from network context.
type Information struct {
	// App groups application/runtime details required for energy computation (e.g., service IP, platform).
	App ApplicationInstance `json:"app"`

	// NE groups network-element details used to attribute/estimate network energy share (correlated via SEN->TEN mapping).
	NE []NEInfo `json:"ne"`
}

type ApplicationInstance struct {
	// IPList is the list of externally reachable endpoints (IPs) of the application's Service.
	IPList []string `json:"ipList"` // Source: Service.status.loadBalancer.ingress[*].ip
	// InfraType identifies the orchestrator/platform of the hosting cluster (e.g., "sylva", "cnis", "os").
	InfraType string `json:"infraType"` // Source: TIMMultiCloudCluster.spec.type
}

type NEInfo struct {
	// InstanceID is the identifier of the Network Element serving the application's location.
	InstanceID string `json:"instanceId"` // Used to query NE telemetry/energy.
	// NetworkID identifies the NE's network/domain (site/region/segment) for scoping energy data.
	NetworkID string `json:"networkId"`
	// VendorID is the NE vendor resolved from the chart mapping.
	VendorID string `json:"vendorId"`
	// Source: ConfigMap keyed by TimApplication.spec.helm.repoUrl + "/" + TimApplication.spec.helm.chartName
	// InfraType classifies the NE (e.g., gNB, UPF) for applying the correct energy model.
	InfraType string `json:"infraType"`
}

type Interface interface {
	// GatherInformation resolves the Information required for energy computation for the given appInstanceID.
	GatherInformation(ctx context.Context, appInstanceID string) (Information, error)
}
