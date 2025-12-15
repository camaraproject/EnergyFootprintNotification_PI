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
package database

import (
	"context"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Job struct {
	JobSpec `bson:",inline"`
	// CalculationTriggered is set when the calculation event has been emitted.
	// Absence or false means it can still be triggered.
	CalculationTriggered bool `bson:"calculationTriggered,omitempty"`
	// NotificationSent is set when a notification event has been sent to the subscriber.
	// Prevents duplicate notifications from multiple DLQ events or concurrent failures.
	NotificationSent bool `bson:"notificationSent,omitempty"`
}

type RequestKind string

const (
	RequestKindEnergyConsumption RequestKind = "energy_consumption"
	RequestKindCarbonFootprint   RequestKind = "carbon_footprint"
)

type JobSpec struct {
	// RequestId Identifier for the request
	RequestId *string `bson:"_id"`

	RequestKind RequestKind `bson:"requestKind"`

	// Service list of Application Instance Identifiers. This are the instances of the applications producing the service under analysis.
	Service []models.AppInstanceId `bson:"service"`

	// SubscriptionRequest The request for creating a event-type event subscription
	SubscriptionRequest models.SubscriptionRequest `bson:"subscriptionRequest"`
	TimePeriod          *models.TimePeriod         `bson:"timePeriod,omitempty"`
}

// JobAppResult represents the result of a single appId within a Job.
// It stores the job reference, app identifier, current status, and the final result payload.
type JobAppResult struct {
	JobAppResultMetadata `bson:",inline"`
	Result               *TaskResult `bson:"result,omitempty"`
}

type JobAppResultMetadata struct {
	// JobID is the identifier of the request this result belongs to.
	JobID string `bson:"jobId"`
	AppID string `bson:"appId"`
	// Total number of NEs considered in the calculation.
	NumberOfTotalNEs int `bson:"numberOfTotalNEs"`
}

// NetworkElementResult holds all results for a single NE in a job/app context.
type NetworkElementResult struct {
	EnergyConsumption  *float64 `bson:"energyConsumption,omitempty"`
	AppInstanceTraffic *float64 `bson:"appInstanceTraffic,omitempty"`
	TotalTraffic       *float64 `bson:"totalTraffic,omitempty"`
}

// TaskResult holds the computed consumption/carbon data for an AppID.
type TaskResult struct {
	AppInstanceEnergyConsumption *float64 `bson:"appInstanceEnergyConsumption"`
	// Maps network element instance ID to all NE results.
	NetworkElements map[string]NetworkElementResult `bson:"networkElements"`
}

type Interface interface {
	// CreateJob inserts a new Job with immutable input and initial status.
	CreateJob(ctx context.Context, r *Job) error

	// GetJob returns a Job by its ID.
	GetJob(ctx context.Context, id string) (*Job, error)

	// SetJobStatus updates the status of a Job by its ID.
	SetJobStatus(ctx context.Context, jobID string, status Status) error

	// CreateOrUpdateNetworkElementResult adds a network element result to a specific JobAppResult. If the JobAppResult does not exist, it creates a new one.
	CreateOrUpdateNetworkElementResult(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, neResult NetworkElementResult) error

	// SetNetworkElementEnergy stores only the energy consumption for a network element without affecting traffic fields.
	SetNetworkElementEnergy(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, energyConsumption float64) error

	// SetNetworkElementTraffic stores only the traffic volume fields for a network element without affecting energy field.
	SetNetworkElementTraffic(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, appInstanceTraffic, totalTraffic float64) error

	// CreateOrUpdateApplicationResult adds the energy consumption result for the service of an application instance to a specific JobAppResult. If the JobAppResult does not exist, it creates a new one.
	CreateOrUpdateApplicationResult(ctx context.Context, creationMetadata JobAppResultMetadata, appInstanceConsumption float64) error

	// GetJobAppResult returns the JobAppResult for a specific AppID within a Job.
	GetJobAppResult(ctx context.Context, jobID, appID string) (*JobAppResult, error)

	// GetAllJobAppResults returns all JobAppResults for a specific JobID.
	GetAllJobAppResults(ctx context.Context, jobID string) ([]JobAppResult, error)

	// TrySetCalculationTriggered atomically sets calculationTriggered=true if it was not already true.
	// Returns true if this call performed the transition (caller may emit event), false if it was already set.
	TrySetCalculationTriggered(ctx context.Context, jobID string) (bool, error)

	// TrySetNotificationSent atomically sets notificationSent=true if it was not already true.
	// Returns true if this call performed the transition (caller may send notification), false if it was already set.
	TrySetNotificationSent(ctx context.Context, jobID string) (bool, error)
}
