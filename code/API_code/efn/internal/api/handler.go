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
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/server"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/event"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/middleware"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/policy"
	servererr "github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/server/error"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/server/request"
)

var _ server.ServerInterface = &handler{}

func New(db database.Interface, pdp policy.Interface) (*handler, error) {
	sender, err := event.NewSender()
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud event sender: %w", err)
	}
	cfg := config.GetConf()
	return &handler{
		events:   sender,
		database: db,
		pdp:      pdp,
		config:   cfg.API,
	}, nil
}

type handler struct {
	database database.Interface
	events   event.Sender
	pdp      policy.Interface
	config   config.API
}

func (h *handler) CalculateCarbonFootprint(c echo.Context, params models.CalculateCarbonFootprintParams) error {
	return h.handleReportCalculation(c, database.RequestKindCarbonFootprint)
}

func (h *handler) CalculateEnergyConsumption(c echo.Context, params models.CalculateEnergyConsumptionParams) error {
	return h.handleReportCalculation(c, database.RequestKindEnergyConsumption)
}

// handleReportCalculation centralizes the shared logic of the two calculate endpoints.
// It binds the common request body, authorizes access to application IDs, creates a job in the database
// and send the CloudEvent.
func (h *handler) handleReportCalculation(c echo.Context, kind database.RequestKind) error {
	log := logger.Get()
	requestID := uuid.New().String()
	ctx := c.Request().Context()

	req, err := request.Bind[models.ReportCreationRequest](c)
	if err != nil {
		msg := "failed to validate request body"
		log.With(zap.Error(err)).Error(msg)
		return servererr.SendFromStatusCode(c, http.StatusBadRequest, msg)
	}

	maxDuration := time.Duration(h.config.MaxTimePeriodDays) * 24 * time.Hour
	now := time.Now()
	oldestAllowedDate := now.Add(-maxDuration)

	// Validate that service array has at least one element
	if len(req.Service) == 0 {
		msg := "service array must contain at least one application instance"
		log.Error(msg)
		return servererr.SendFromStatusCode(c, http.StatusBadRequest, msg)
	}

	// Validate that timePeriod is provided; if not, set default
	if req.TimePeriod == nil {
		req.TimePeriod = &models.TimePeriod{
			StartDate: oldestAllowedDate,
			EndDate:   &now,
		}
	}

	// Validate that endDate is after startDate if endDate is present
	if req.TimePeriod.EndDate != nil {
		if !req.TimePeriod.EndDate.After(req.TimePeriod.StartDate) {
			msg := "endDate must be after startDate in timePeriod"
			log.Error(msg)
			return servererr.SendFromStatusCode(c, http.StatusBadRequest, msg)
		}
	}

	// Validate that subscription event type matches the endpoint
	if len(req.SubscriptionRequest.Types) > 0 {
		expectedType := getExpectedEventType(kind)
		actualType := string(req.SubscriptionRequest.Types[0])
		if actualType != expectedType {
			msg := fmt.Sprintf("subscription event type '%s' does not match endpoint (expected '%s')", actualType, expectedType)
			log.Error(msg)
			return servererr.SendFromStatusCode(c, http.StatusBadRequest, msg)
		}
	}

	// Validate that startDate and endDate are not in the future
	if req.TimePeriod.StartDate.After(now) {
		msg := "startDate cannot be in the future"
		log.Error(msg)
		return servererr.SendFromStatusCodeWithCode(c, http.StatusBadRequest, "OUT_OF_RANGE", msg)
	}
	if req.TimePeriod.EndDate != nil && req.TimePeriod.EndDate.After(now) {
		msg := "endDate cannot be in the future"
		log.Error(msg)
		return servererr.SendFromStatusCodeWithCode(c, http.StatusBadRequest, "OUT_OF_RANGE", msg)
	}

	if req.TimePeriod.StartDate.Before(oldestAllowedDate) {
		msg := fmt.Sprintf("startDate cannot be older than %d days", h.config.MaxTimePeriodDays)
		log.Error(msg)
		return servererr.SendFromStatusCodeWithCode(c, http.StatusBadRequest, "OUT_OF_RANGE", msg)
	}
	if req.TimePeriod.EndDate != nil && req.TimePeriod.EndDate.Before(oldestAllowedDate) {
		msg := fmt.Sprintf("endDate cannot be older than %d days", h.config.MaxTimePeriodDays)
		log.Error(msg)
		return servererr.SendFromStatusCodeWithCode(c, http.StatusBadRequest, "OUT_OF_RANGE", msg)
	}

	// Validate protocol limitation (only HTTP supported now)
	if err = req.SubscriptionRequest.ValidateProtocol(); err != nil {
		log.With(zap.Error(err)).Warn("unsupported subscription protocol")
		return servererr.SendFromStatusCode(c, http.StatusNotImplemented, err.Error())
	}

	// Validate sink credential limitation (only ACCESSTOKEN supported now)
	if cred := req.SubscriptionRequest.SinkCredential; cred != nil {
		if err = cred.Validate(); err != nil {
			log.With(zap.Error(err)).Warn("sink credential validation failed")
			return servererr.SendFromStatusCode(c, http.StatusNotImplemented, err.Error())
		}
	}

	// Warn if initialEvent is set to true (has no effect for this API)
	if req.SubscriptionRequest.Config.InitialEvent != nil && *req.SubscriptionRequest.Config.InitialEvent {
		log.Warn("initialEvent is set to true but has no effect for this API")
	}

	appIds := make([]string, len(req.Service))
	for i, id := range req.Service {
		appIds[i] = id.String()
	}

	if err = h.pdp.HasAccessToApplicationIDs(ctx, middleware.CtxSub(ctx), appIds); err != nil {
		msg := "failed to authorize application IDs"
		log.With(zap.Error(err)).Error(msg)
		return servererr.SendFromStatusCode(c, http.StatusUnauthorized, err.Error())
	}

	req.RequestId = &requestID
	if err = h.database.CreateJob(ctx, newJob(*req, kind)); err != nil {
		log.With(zap.Error(err)).Error("failed to create job")
		return servererr.Send(c, err)
	}

	for _, appInstanceID := range appIds {
		eventId := event.EventIDForApp(requestID, appInstanceID)
		if err = h.events.Send(ctx, eventId, event.EventTypeGatherInfoRequested, event.SourceEFNAPI, event.NewGatherInfoData(requestID, appInstanceID)); err != nil {
			msg := "failed to send cloud event"
			log.With(zap.Error(err), zap.String("Event ID", eventId)).Error(msg)
			return servererr.SendFromStatusCode(c, http.StatusInternalServerError, "failed to send event")
		}
	}
	return c.JSON(http.StatusCreated, req)
}

func newJob(req models.ReportCreationRequest, kind database.RequestKind) *database.Job {
	return &database.Job{
		JobSpec: database.JobSpec{
			RequestId:           req.RequestId,
			RequestKind:         kind,
			Service:             req.Service,
			SubscriptionRequest: req.SubscriptionRequest,
			TimePeriod:          req.TimePeriod,
		},
	}
}

func getExpectedEventType(kind database.RequestKind) string {
	switch kind {
	case database.RequestKindEnergyConsumption:
		return "org.camaraproject.energy-footprint-notification.v1.energy"
	case database.RequestKindCarbonFootprint:
		return "org.camaraproject.energy-footprint-notification.v1.carbon-footprint"
	default:
		return ""
	}
}
