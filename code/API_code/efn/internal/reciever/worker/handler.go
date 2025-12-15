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
package worker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	cloudevent "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/calculator"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/cloudobservability"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/event"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/orchestrator"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/trafficvolume"
)

// Handler processes JobRequested events for the Worker service.
type Handler struct {
	calculator         calculator.Interface
	cloudObservability cloudobservability.Interface
	trafficVolume      trafficvolume.Interface
	database           database.Interface
	events             event.Sender
	orchestrator       orchestrator.Interface
}

func NewHandler(db database.Interface, orch orchestrator.Interface) (*Handler, error) {
	sender, err := event.NewSender()
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud event sender: %w", err)
	}
	var cloudObs cloudobservability.Interface
	// Check for configurable client first
	if os.Getenv("CLIENT_TYPE") == "configurable" {
		cloudObs, err = cloudobservability.NewConfigurableClient()
	} else if os.Getenv("CLOUDOBS_FAIL_NE") == "true" || os.Getenv("CLOUDOBS_FAIL_THROTTLE") == "true" {
		cloudObs, err = cloudobservability.NewErrorDummyClient()
	} else {
		cloudObs, err = cloudobservability.NewDummyClient()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud observability client: %w", err)
	}

	var tv trafficvolume.Interface
	// Check for configurable traffic volume client
	if os.Getenv("TRAFFIC_CLIENT_TYPE") == "configurable" {
		tv, err = trafficvolume.NewConfigurableClient()
	} else {
		tv, err = trafficvolume.NewDummyClient()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic volume client: %w", err)
	}

	// Read CO2 conversion factor from environment (tCO2e per kWh)
	// Default to 0.00035 if not provided or invalid
	var carbonFactor float64 = 0.00035
	if val := os.Getenv("CARBON_FACTOR_TCO2E_PER_KWH"); val != "" {
		if f, perr := strconv.ParseFloat(val, 64); perr == nil && f > 0 {
			carbonFactor = f
		}
	}

	return &Handler{
		calculator:         calculator.NewSimpleClient(carbonFactor),
		cloudObservability: cloudObs,
		trafficVolume:      tv,
		database:           db,
		events:             sender,
		orchestrator:       orch,
	}, nil
}

func (h *Handler) Handle(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.With(zap.String("type", e.Type()), zap.String("source", e.Source())).Debug("Received event")

	switch e.Type() {
	case event.EventTypeGatherInfoRequested.String():
		return h.handleGatherInfoRequested(ctx, e)
	case event.EventTypeAppConsumptionRequested.String():
		return h.handleAppConsumptionRequested(ctx, e)
	case event.EventTypeNetworkElementEnergyRequested.String():
		return h.handleNetworkElementEnergyRequested(ctx, e)
	case event.EventTypeNetworkElementTrafficRequested.String():
		return h.handleNetworkElementTrafficRequested(ctx, e)
	case event.EventTypeCalculationRequested.String():
		return h.handleCalculationRequested(ctx, e)
	default:
		log.Error("unknown event: " + e.Type())
		return nil, nil
	}
}

// handleCalculationRequested processes the calculation event after all data is gathered.
func (h *Handler) handleCalculationRequested(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.Debug("Handling Calculation Requested event")

	requestID := e.ID()
	log = log.With(zap.String("request/job id", requestID))

	jobID := requestID
	appResults, err := h.database.GetAllJobAppResults(ctx, jobID)
	if err != nil {
		msg := "Failed to fetch all job app results for calculation"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	log.With(zap.Int("numAppResults", len(appResults))).Debug("Fetched all JobAppResults for calculation")

	// Retrieve job to determine request kind
	job, err := h.database.GetJob(ctx, jobID)
	if err != nil {
		msg := "Failed to fetch job for calculation"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// Calculate based on request kind
	var result *float64
	switch job.RequestKind {
	case database.RequestKindEnergyConsumption:
		result, err = h.calculator.CalculateEnergyConsumption(ctx, appResults)
		if err != nil {
			msg := "Failed to calculate energy consumption"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.With(zap.Float64("energyConsumption", *result)).Debug("Successfully calculated energy consumption")
	case database.RequestKindCarbonFootprint:
		result, err = h.calculator.CalculateCarbonFootprint(ctx, appResults)
		if err != nil {
			msg := "Failed to calculate carbon footprint"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.With(zap.Float64("carbonFootprint", *result)).Debug("Successfully calculated carbon footprint")
	default:
		msg := "Unknown request kind, defaulting to energy consumption"
		log.Warn(msg, zap.String("requestKind", string(job.RequestKind)))
		result, err = h.calculator.CalculateEnergyConsumption(ctx, appResults)
		if err != nil {
			msg := "Failed to calculate energy consumption"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.With(zap.Float64("energyConsumption", *result)).Debug("Successfully calculated energy consumption")
	}

	log.Info("Calculation completed and sending notification event", zap.String("jobID", jobID))
	notificationData := event.NewNotificationRequestedData(requestID, *result)
	if err := h.events.Send(ctx, requestID, event.EventTypeNotificationRequested, event.SourceEFNWorker, notificationData); err != nil {
		log.With(zap.Error(err)).Error("Failed to send NotificationRequested event")
		return nil, fmt.Errorf("failed to send NotificationRequested event: %w", err)
	}
	return nil, nil
}

func (h *Handler) handleAppConsumptionRequested(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.Debug("Handling App Consumption Requested")

	data := event.AppConsumptionData{}
	err := e.DataAs(&data)
	if err != nil {
		msg := "Failed to parse event data"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	log.With(zap.Any("data", data)).Debug("App Consumption Data parsed from event")
	log = log.With(zap.String("request/job id", data.RequestID), zap.String("appInstanceID", data.ApplicationInstanceID))

	consumption, err := h.cloudObservability.RetrieveAppEnergyConsumption(ctx, data.ApplicationInstanceID, data.TimePeriod, data.AppInfraType)
	if err != nil {
		if cloudobservability.IsThrottlingError(err) {
			log.With(zap.Error(err)).Warn("Throttling error retrieving app energy consumption")
			return nil, fmt.Errorf("throttling error: %w", err)
		}
		log.With(zap.Error(err)).Error("Permanent error retrieving app energy consumption")

		notificationData := event.NewNotificationErrorRequestedData(
			data.RequestID,
			http.StatusInternalServerError,
			"Failed to retrieve app energy consumption",
		)
		if sendErr := h.events.Send(ctx, data.RequestID, event.EventTypeNotificationErrorRequested, event.SourceEFNWorker, notificationData); sendErr != nil {
			log.With(zap.Error(sendErr)).Error("Failed to send error notification")
			return nil, fmt.Errorf("failed to send error notification: %w", sendErr)
		}
		log.Info("Sent error notification for failed app consumption retrieval")
		return nil, nil
	}
	log.With(zap.Float64("consumption", *consumption)).Debug("Successfully retrieved app energy consumption")

	// Store the result in the database
	creationMetadata := database.JobAppResultMetadata{
		JobID:            data.RequestID,
		AppID:            data.ApplicationInstanceID,
		NumberOfTotalNEs: data.NumberOfTotalNEs,
	}
	if err := h.database.CreateOrUpdateApplicationResult(ctx, creationMetadata, *consumption); err != nil {
		msg := "Failed to store energy consumption for the application instance in database"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	isAllDataGathered, err := h.isAllDataGathered(ctx, data.RequestID)
	if err != nil {
		msg := "Failed to verify if all data is gathered"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	log.With(zap.Bool("isAllDataGathered", isAllDataGathered)).Debug("App consumption: checked if all data gathered")

	if isAllDataGathered {
		// Send Calculation Requested event (previously returned from handler - now explicitly sent since main.go doesn't forward returned events)
		log.Info("Sending CalculationRequested event after app consumption")
		if err := h.events.Send(ctx, data.RequestID, event.EventTypeCalculationRequested, event.SourceEFNWorker, event.NewCalculationRequestedData()); err != nil {
			msg := "Failed to send CalculationRequested event"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.Debug("Successfully sent CalculationRequested event")
	}

	return nil, nil
}

func (h *Handler) handleGatherInfoRequested(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	data := event.GatherInfoData{}
	err := e.DataAs(&data)
	if err != nil {
		msg := "Failed to parse event data"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	jobID := data.RequestID
	appInstanceID := data.ApplicationInstanceID
	log = log.With(zap.String("requestID", jobID), zap.String("appInstanceID", appInstanceID))

	job, err := h.database.GetJob(ctx, jobID)
	if err != nil {
		log.With(zap.Error(err), zap.String("requestID", jobID)).Error("failed to read DB Job")
		return nil, fmt.Errorf("failed to read job "+jobID+" from DB: %w", err)
	}
	log.With(zap.Any("Job", job)).Debug("Successfully gotten job from DB")

	info, err := h.orchestrator.GatherInformation(ctx, appInstanceID)
	if err != nil {
		msg := "Failed to gather application instance information with ID " + string(appInstanceID)
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf(msg+": %w", err)
	}
	log.With(zap.Any("Info", info)).Debug("Successfully gathered info from orchestrator")

	// Send App Consumption Requested event for the application instance
	eventId := event.EventIDForApp(jobID, appInstanceID)
	numberOfTotalNEs := len(info.NE)
	eventData := event.NewAppConsumptionData(
		jobID,
		appInstanceID,
		job.TimePeriod,
		info.App.InfraType,
		numberOfTotalNEs,
	)
	if err = h.events.Send(ctx, eventId, event.EventTypeAppConsumptionRequested, event.SourceEFNWorker, eventData); err != nil {
		msg := "Failed to send cloud event to get app consumption"
		log.With(zap.Error(err), zap.String("Event ID", eventId)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// Build array of all network elements
	networkElements := make([]event.NetworkElementInfo, 0, len(info.NE))
	for _, neInfo := range info.NE {
		networkElements = append(networkElements, event.NetworkElementInfo{
			NEInstanceID: neInfo.InstanceID,
			VendorID:     neInfo.VendorID,
			NetworkID:    neInfo.NetworkID,
			NEInfraType:  neInfo.InfraType,
		})
	}

	// Send individual Network Element Energy Requested events for each NE
	for _, neInfo := range info.NE {
		eventId := event.EventIDForNE(jobID, appInstanceID, neInfo.InstanceID)
		neEnergyEventData := event.NewNetworkElementEnergyData(
			jobID,
			appInstanceID,
			neInfo.InstanceID,
			neInfo.InfraType,
			job.TimePeriod,
			numberOfTotalNEs,
		)
		if err = h.events.Send(ctx, eventId, event.EventTypeNetworkElementEnergyRequested, event.SourceEFNWorker, neEnergyEventData); err != nil {
			msg := "Failed to send cloud event to get network element energy consumption"
			log.With(zap.Error(err), zap.String("Event ID", eventId), zap.String("NE ID", neInfo.InstanceID)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
	}

	// Send single Network Element Traffic Requested event with all network elements
	eventId = event.EventIDForTraffic(jobID, appInstanceID)
	neTrafficEventData := event.NewNetworkElementTrafficData(
		jobID,
		appInstanceID,
		info.App.IPList,
		job.TimePeriod,
		networkElements,
	)
	if err = h.events.Send(ctx, eventId, event.EventTypeNetworkElementTrafficRequested, event.SourceEFNWorker, neTrafficEventData); err != nil {
		msg := "Failed to send cloud event to get network element traffic volume"
		log.With(zap.Error(err), zap.String("Event ID", eventId), zap.Int("NE Count", len(networkElements))).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	return nil, nil
}

func (h *Handler) handleNetworkElementEnergyRequested(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.Debug("Handling Network Element Energy Requested")

	data := event.NetworkElementEnergyData{}
	err := e.DataAs(&data)
	if err != nil {
		msg := "Failed to parse event data"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	log.With(
		zap.String("requestID", data.RequestID),
		zap.String("appInstanceID", data.ApplicationInstanceID),
		zap.String("neInstanceID", data.NEInstanceID),
	).Debug("Network Element Energy Data parsed from event")

	log = log.With(
		zap.String("request/job id", data.RequestID),
		zap.String("appInstanceID", data.ApplicationInstanceID),
		zap.String("neInstanceID", data.NEInstanceID),
	)

	consumption, err := h.cloudObservability.RetrieveNetworkElementEnergyConsumption(ctx, data.ApplicationInstanceID, data.TimePeriod, data.NEInfraType)
	if err != nil {
		if cloudobservability.IsThrottlingError(err) {
			log.With(zap.Error(err)).Warn("Throttling error retrieving network element energy consumption")
			return nil, fmt.Errorf("throttling error for NE %s: %w", data.NEInstanceID, err)
		}
		log.With(zap.Error(err)).Error("Permanent error retrieving network element energy consumption")

		notificationData := event.NewNotificationErrorRequestedData(
			data.RequestID,
			http.StatusInternalServerError,
			"Failed to retrieve network element energy consumption",
		)
		if sendErr := h.events.Send(ctx, data.RequestID, event.EventTypeNotificationErrorRequested, event.SourceEFNWorker, notificationData); sendErr != nil {
			log.With(zap.Error(sendErr)).Error("Failed to send error notification")
			return nil, fmt.Errorf("failed to send error notification: %w", sendErr)
		}
		log.Info("Sent error notification for failed network element energy retrieval")
		return nil, nil
	}
	log.With(zap.Float64("consumption", *consumption)).Debug("Successfully retrieved network element energy consumption")

	// Store only the energy consumption in the database
	creationMetadata := database.JobAppResultMetadata{
		JobID:            data.RequestID,
		AppID:            data.ApplicationInstanceID,
		NumberOfTotalNEs: data.NumberOfTotalNEs,
	}
	if err := h.database.SetNetworkElementEnergy(ctx, creationMetadata, data.NEInstanceID, *consumption); err != nil {
		msg := "Failed to store energy consumption for the network element in database"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s for NE %s: %w", msg, data.NEInstanceID, err)
	}
	log.Debug("Successfully stored network element energy in database")

	isAllDataGathered, err := h.isAllDataGathered(ctx, data.RequestID)
	if err != nil {
		msg := "Failed to verify if all data is gathered"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	log.With(zap.Bool("isAllDataGathered", isAllDataGathered)).Debug("NE energy: checked if all data gathered")

	if isAllDataGathered {
		// Send Calculation Requested event
		log.Info("Sending CalculationRequested event after NE energy")
		if err := h.events.Send(ctx, data.RequestID, event.EventTypeCalculationRequested, event.SourceEFNWorker, event.NewCalculationRequestedData()); err != nil {
			msg := "Failed to send CalculationRequested event"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.Debug("Successfully sent CalculationRequested event")
	}
	return nil, nil
}

func (h *Handler) handleNetworkElementTrafficRequested(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.Debug("Handling Network Element Traffic Requested")

	data := event.NetworkElementTrafficData{}
	err := e.DataAs(&data)
	if err != nil {
		msg := "Failed to parse event data"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	log.With(
		zap.String("requestID", data.RequestID),
		zap.String("appInstanceID", data.ApplicationInstanceID),
		zap.Int("neCount", len(data.NetworkElements)),
		zap.Int("ipCount", len(data.AppInstanceIPList)),
	).Debug("Network Element Traffic Data parsed from event")

	log = log.With(
		zap.String("request/job id", data.RequestID),
		zap.String("appInstanceID", data.ApplicationInstanceID),
	)

	// Build list of network elements for Traffic Volume API call
	tvNetworkElements := make([]trafficvolume.NetworkElement, 0, len(data.NetworkElements))
	for _, neInfo := range data.NetworkElements {
		tvNetworkElements = append(tvNetworkElements, trafficvolume.NetworkElement{
			VendorIdentifier: neInfo.VendorID,
			NEIdentifier:     neInfo.NEInstanceID,
		})
	}

	// Retrieve traffic volumes for all network elements in one call
	trafficVolumes, err := h.trafficVolume.RetrieveTrafficVolumes(ctx, data.AppInstanceIPList, tvNetworkElements, data.TimePeriod)
	if err != nil {
		msg := "Failed to retrieve traffic volumes from Traffic Volume API"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	log.With(zap.Int("measureCount", len(trafficVolumes.TrafficVolumeMeasureList))).Debug("Successfully retrieved traffic volumes from Traffic Volume API")

	// Create a map for quick lookup of traffic volumes by NE identifier
	trafficVolumeMap := make(map[string]trafficvolume.TrafficVolumeMeasure)
	for _, measure := range trafficVolumes.TrafficVolumeMeasureList {
		trafficVolumeMap[measure.NetworkElement.NEIdentifier] = measure
	}

	// Process and store traffic for each network element
	for _, neInfo := range data.NetworkElements {
		log.With(
			zap.String("neInstanceID", neInfo.NEInstanceID),
			zap.String("vendorID", neInfo.VendorID),
		).Debug("Processing network element traffic")

		// Get traffic volumes from the map
		trafficVolume, ok := trafficVolumeMap[neInfo.NEInstanceID]
		if !ok {
			msg := "Traffic volume not found in Traffic Volume API response"
			log.With(zap.String("neInstanceID", neInfo.NEInstanceID)).Error(msg)
			return nil, fmt.Errorf("%s for NE %s", msg, neInfo.NEInstanceID)
		}

		traffic := trafficVolume.TrafficVolumeIP
		totalNETraffic := trafficVolume.TrafficVolumeAll
		log.With(
			zap.Float64("traffic", traffic),
			zap.Float64("totalTraffic", totalNETraffic),
			zap.String("neInstanceID", neInfo.NEInstanceID),
		).Debug("Retrieved traffic volumes for network element")

		// Store only the traffic volume fields in the database
		creationMetadata := database.JobAppResultMetadata{
			JobID:            data.RequestID,
			AppID:            data.ApplicationInstanceID,
			NumberOfTotalNEs: len(data.NetworkElements),
		}
		if err := h.database.SetNetworkElementTraffic(ctx, creationMetadata, neInfo.NEInstanceID, traffic, totalNETraffic); err != nil {
			msg := "Failed to store traffic for the network element in database"
			log.With(zap.Error(err), zap.String("neInstanceID", neInfo.NEInstanceID)).Error(msg)
			return nil, fmt.Errorf("%s for NE %s: %w", msg, neInfo.NEInstanceID, err)
		}
		log.With(zap.String("neInstanceID", neInfo.NEInstanceID)).Debug("Successfully stored network element traffic in database")
	}

	isAllDataGathered, err := h.isAllDataGathered(ctx, data.RequestID)
	if err != nil {
		msg := "Failed to verify if all data is gathered"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	log.With(zap.Bool("isAllDataGathered", isAllDataGathered)).Debug("NE traffic: checked if all data gathered")

	if isAllDataGathered {
		// Send Calculation Requested event
		log.Info("Sending CalculationRequested event after NE traffic")
		if err := h.events.Send(ctx, data.RequestID, event.EventTypeCalculationRequested, event.SourceEFNWorker, event.NewCalculationRequestedData()); err != nil {
			msg := "Failed to send CalculationRequested event"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.Debug("Successfully sent CalculationRequested event")
	}
	return nil, nil
}

func (h *Handler) isAllDataGathered(ctx context.Context, requestID string) (bool, error) {
	log := logger.Get().With(zap.String("requestID", requestID))

	// Fetch job to know expected number of app instances
	job, err := h.database.GetJob(ctx, requestID)
	if err != nil {
		return false, fmt.Errorf("failed to read job %s: %w", requestID, err)
	}

	results, err := h.database.GetAllJobAppResults(ctx, requestID)
	if err != nil {
		return false, fmt.Errorf("failed to get all job app results for requestID %s: %w", requestID, err)
	}

	// If it hasn't yet created results for every app instance, it's not done.
	expectedApps := len(job.JobSpec.Service)
	if len(results) != expectedApps {
		log.With(zap.Int("expected", expectedApps), zap.Int("actual", len(results))).Debug("Not all app results present yet")
		return false, nil
	}

	// Check each app instance is fully populated.
	for i, appResult := range results {
		result := appResult.Result
		if result == nil || result.AppInstanceEnergyConsumption == nil {
			log.With(zap.Int("appIndex", i), zap.String("appID", appResult.AppID)).Debug("App result incomplete: missing result or consumption")
			return false, nil
		}
		if appResult.NumberOfTotalNEs != len(result.NetworkElements) {
			log.With(
				zap.Int("appIndex", i),
				zap.String("appID", appResult.AppID),
				zap.Int("expectedNEs", appResult.NumberOfTotalNEs),
				zap.Int("actualNEs", len(result.NetworkElements)),
			).Debug("App result incomplete: NE count mismatch")
			return false, nil
		}
		for neID, neResult := range result.NetworkElements {
			if neResult.EnergyConsumption == nil || neResult.AppInstanceTraffic == nil || neResult.TotalTraffic == nil {
				log.With(
					zap.Int("appIndex", i),
					zap.String("appID", appResult.AppID),
					zap.String("neID", neID),
				).Debug("NE result incomplete: missing field")
				return false, nil
			}
		}
	}

	log.Debug("All data gathered! Attempting to set calculationTriggered flag")

	// All data gathered; attempt cross-pod atomic flag set.
	set, err := h.database.TrySetCalculationTriggered(ctx, requestID)
	if err != nil {
		return false, fmt.Errorf("failed to set calculationTriggered for %s: %w", requestID, err)
	}
	if !set {
		// Another pod already triggered calculation.
		log.Debug("CalculationTriggered flag was already set by another pod/event")
		return false, nil
	}
	log.Debug("Successfully set calculationTriggered flag - this pod will emit event")
	return true, nil
}

func (h *Handler) HandleDLQEvent(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.With(zap.String("eventType", e.Type()), zap.String("eventID", e.ID())).Info("DLQ event received after broker retries exhausted")

	// All event types in our system have RequestID field, so we can use a generic approach
	var eventData struct {
		RequestID string `json:"requestId"`
	}
	if err := e.DataAs(&eventData); err != nil {
		log.With(zap.Error(err)).Error("Failed to parse requestId from DLQ event data")
		return nil, err
	}

	requestID := eventData.RequestID
	log = log.With(zap.String("requestID", requestID))
	log.Debug("Extracted request ID from DLQ event")
	notificationData := event.NewNotificationErrorRequestedData(
		requestID,
		http.StatusInternalServerError,
		"Event processing failed after multiple retries",
	)
	if err := h.events.Send(ctx, requestID, event.EventTypeNotificationErrorRequested, event.SourceEFNWorker, notificationData); err != nil {
		log.With(zap.Error(err)).Error("Failed to send error notification from DLQ")
		return nil, err
	}
	log.Info("Successfully sent error notification event from DLQ")
	return nil, nil
}
