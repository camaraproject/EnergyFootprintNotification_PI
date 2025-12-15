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

type EventType string

const (
	// EventTypeGatherInfoRequested is sent by the API to the worker to gather info for a requestID and applicationInstanceID.
	EventTypeGatherInfoRequested EventType = "it.tim.efn.gatherinfo.requested"

	// EventTypeAppConsumptionRequested is sent by the worker to get energy consumption for an applicationInstanceID.
	EventTypeAppConsumptionRequested EventType = "it.tim.efn.app.consumption.requested"

	// EventTypeNetworkElementEnergyRequested is sent by the worker to get energy consumption for a single network element from CloudObservability.
	EventTypeNetworkElementEnergyRequested EventType = "it.tim.efn.networkelement.energy.requested"

	// EventTypeNetworkElementTrafficRequested is sent by the worker to get traffic volume info for all network elements from TrafficVolumeAPI (batch).
	EventTypeNetworkElementTrafficRequested EventType = "it.tim.efn.networkelement.traffic.requested"

	// EventTypeNotificationRequested is sent by the EFN Worker when the calculation has completed successfully.
	EventTypeNotificationRequested EventType = "it.tim.efn.notification.requested"

	// EventTypeNotificationErrorRequested is sent by the EFN Worker when an error occurs during processing.
	EventTypeNotificationErrorRequested EventType = "it.tim.efn.notification.error.requested"

	// EventTypeNotificationSent is sent by the EFN Notify service when a notification has been sent.
	EventTypeNotificationSent EventType = "it.tim.efn.notification.sent"

	// EventTypeCalculationRequested is sent by the worker to itself to trigger calculation after all values are gathered.
	EventTypeCalculationRequested EventType = "it.tim.efn.calculation.requested"
)

func (s EventType) String() string {
	return string(s)
}

type Source string

const (
	// SourceEFNAPI is the CloudEvents source for the EFN API service.
	SourceEFNAPI Source = "urn:tim:efn-api"

	// SourceEFNWorker is the CloudEvents source for the EFN Worker service.
	SourceEFNWorker Source = "urn:tim:efn-worker"

	// SourceEFNNotify is the CloudEvents source for the EFN Notify service.
	SourceEFNNotify Source = "urn:tim:efn-notify"
)

func (s Source) String() string {
	return string(s)
}

// Extension key used for DLQ retry counting. Must be lowercase to survive CloudEvents HTTP binary mapping.
// CloudEvents recommendation: use lower-case attribute names; binary transport normalizes headers to lower-case.
const RetryExtensionKey = "efnretrycount"
