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
package event

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

// GatherInfoData is the CloudEvent payload for the GatherInfo event at
// application-instance scope.
type GatherInfoData struct {
	RequestID             string `json:"requestId"`
	ApplicationInstanceID string `json:"applicationInstanceId"`
}

// AppConsumptionData is the CloudEvent payload for app-level energy consumption requests.
type AppConsumptionData struct {
	RequestID             string             `json:"requestId"`
	ApplicationInstanceID string             `json:"applicationInstanceId"`
	TimePeriod            *models.TimePeriod `json:"timePeriod"`
	AppInfraType          string             `json:"appInfraType"`
	NumberOfTotalNEs      int                `json:"numberOfTotalNEs"`
}

// NetworkElementInfo contains information about a single network element.
type NetworkElementInfo struct {
	NEInstanceID string `json:"neInstanceId"`
	VendorID     string `json:"vendorId"`
	NetworkID    string `json:"networkId"`
	NEInfraType  string `json:"neInfraType"`
}

// NetworkElementEnergyData is the CloudEvent payload for a single network element energy consumption request.
type NetworkElementEnergyData struct {
	RequestID             string             `json:"requestId"`
	ApplicationInstanceID string             `json:"applicationInstanceId"`
	NEInstanceID          string             `json:"neInstanceId"`
	NEInfraType           string             `json:"neInfraType"`
	TimePeriod            *models.TimePeriod `json:"timePeriod"`
	NumberOfTotalNEs      int                `json:"numberOfTotalNEs"`
}

// NetworkElementTrafficData is the CloudEvent payload for batch traffic volume requests to TrafficVolumeAPI.
type NetworkElementTrafficData struct {
	RequestID             string               `json:"requestId"`
	ApplicationInstanceID string               `json:"applicationInstanceId"`
	AppInstanceIPList     []string             `json:"appInstanceIpList"`
	TimePeriod            *models.TimePeriod   `json:"timePeriod"`
	NetworkElements       []NetworkElementInfo `json:"networkElements"`
}

// NotificationRequestedData is the CloudEvent payload for notification requested events.
type NotificationRequestedData struct {
	RequestID string  `json:"requestId"`
	Result    float64 `json:"result"`
}

// NotificationErrorRequestedData is the CloudEvent payload for error notification events.
type NotificationErrorRequestedData struct {
	RequestID string `json:"requestId"`
	models.ErrorInfo
}

// CalculationRequestedData is the CloudEvent payload for the CalculationRequested event.
// The data structure is intentionally empty; all data is retrieved from the database.
type CalculationRequestedData struct{}

// NewGatherInfoData returns the payload for a GatherInfo event scoped
// to an application instance.
func NewGatherInfoData(requestID, appInstanceID string) GatherInfoData {
	return GatherInfoData{
		RequestID:             requestID,
		ApplicationInstanceID: appInstanceID,
	}
}

// NewAppConsumptionData returns the payload for an app-level consumption event.
func NewAppConsumptionData(requestID, appInstanceID string, timePeriod *models.TimePeriod, appInfraType string, numberOfTotalNEs int) AppConsumptionData {
	return AppConsumptionData{
		RequestID:             requestID,
		ApplicationInstanceID: appInstanceID,
		TimePeriod:            timePeriod,
		AppInfraType:          appInfraType,
		NumberOfTotalNEs:      numberOfTotalNEs,
	}
}

// NewNotificationRequestedData returns the payload for a notification requested event.
func NewNotificationRequestedData(requestID string, result float64) NotificationRequestedData {
	return NotificationRequestedData{
		RequestID: requestID,
		Result:    result,
	}
}

// NewNotificationErrorRequestedData returns the payload for an error notification event.
func NewNotificationErrorRequestedData(requestID string, status int, message string) NotificationErrorRequestedData {
	return NotificationErrorRequestedData{
		RequestID: requestID,
		ErrorInfo: models.ErrorInfo{
			Status:  status,
			Code:    http.StatusText(status),
			Message: message,
		},
	}
}

// NewNetworkElementEnergyData returns the payload for a NetworkElementEnergyRequested event.
func NewNetworkElementEnergyData(
	requestID, appInstanceID, neInstanceID, neInfraType string,
	timePeriod *models.TimePeriod,
	numberOfTotalNEs int,
) NetworkElementEnergyData {
	return NetworkElementEnergyData{
		RequestID:             requestID,
		ApplicationInstanceID: appInstanceID,
		NEInstanceID:          neInstanceID,
		NEInfraType:           neInfraType,
		TimePeriod:            timePeriod,
		NumberOfTotalNEs:      numberOfTotalNEs,
	}
}

// NewNetworkElementTrafficData returns the payload for a NetworkElementTrafficRequested event.
func NewNetworkElementTrafficData(
	requestID, appInstanceID string,
	appInstanceIPList []string,
	timePeriod *models.TimePeriod,
	networkElements []NetworkElementInfo,
) NetworkElementTrafficData {
	return NetworkElementTrafficData{
		RequestID:             requestID,
		ApplicationInstanceID: appInstanceID,
		AppInstanceIPList:     appInstanceIPList,
		TimePeriod:            timePeriod,
		NetworkElements:       networkElements,
	}
}

// NewCalculationRequestedData returns the payload for a CalculationRequested event.
// The data structure is intentionally empty; all data is retrieved from the database.
func NewCalculationRequestedData() CalculationRequestedData {
	return CalculationRequestedData{}
}

// EventIDForApp returns a deterministic UUIDv5 from requestID and appInstanceID.
func EventIDForApp(requestID, appInstanceID string) string {
	baseNS := uuid.NewSHA1(uuid.NameSpaceURL, []byte("camara-efn-api:event-id"))
	name := requestID + "\x00" + appInstanceID
	return uuid.NewSHA1(baseNS, []byte(name)).String()
}

// EventIDForNE returns a deterministic UUIDv5 for a single network element energy request.
func EventIDForNE(requestID, appInstanceID, neInstanceID string) string {
	baseNS := uuid.NewSHA1(uuid.NameSpaceURL, []byte("camara-efn-api:event-id:ne"))
	name := requestID + "\x00" + appInstanceID + "\x00" + neInstanceID
	return uuid.NewSHA1(baseNS, []byte(name)).String()
}

// EventIDForTraffic returns a deterministic UUIDv5 for traffic volume batch requests.
func EventIDForTraffic(requestID, appInstanceID string) string {
	baseNS := uuid.NewSHA1(uuid.NameSpaceURL, []byte("camara-efn-api:event-id:traffic"))
	name := requestID + "\x00" + appInstanceID
	return uuid.NewSHA1(baseNS, []byte(name)).String()
}
