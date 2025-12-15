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
package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudevent "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/api/models"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/event"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
)

// Handler processes NotificationRequested events for the Notification service.
type Handler struct {
	db     database.Interface
	config config.HTTP
}

func NewHandler(db database.Interface, httpConfig config.HTTP) *Handler {
	return &Handler{
		db:     db,
		config: httpConfig,
	}
}

// getHTTPClient returns an HTTP client configured based on the sink URL.
// For internal cluster services (*.svc.cluster.local), TLS verification
// can be skipped if configured via HTTP_INSECURE_SKIP_VERIFY.
func (h *Handler) getHTTPClient(sinkURL string) *http.Client {
	// Check if the sink is an internal cluster service
	if isInternalClusterService(sinkURL) {
		logger.Get().Debug("Detected internal cluster service", zap.String("sink", sinkURL), zap.Bool("insecureSkipVerify", h.config.InsecureSkipVerify))
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: h.config.InsecureSkipVerify,
				},
			},
		}
	}

	// For external services, use standard client with system CA pool
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// isInternalClusterService checks if the URL points to an internal Kubernetes service.
// Returns true for URLs with hostnames ending in .svc.cluster.local or just .svc
func isInternalClusterService(sinkURL string) bool {
	u, err := url.Parse(sinkURL)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(u.Hostname())
	return strings.HasSuffix(hostname, ".svc.cluster.local") ||
		strings.HasSuffix(hostname, ".svc") ||
		strings.Contains(hostname, ".svc.")
}

// Handle receives the internal NotificationRequested or NotificationErrorRequested event and delivers a CAMARA-compliant
// CloudEvent to the subscriber sink. It then emits an internal NotificationSent event.
func (h *Handler) Handle(ctx context.Context, e cloudevent.Event) (*cloudevent.Event, error) {
	log := logger.Get()
	log.With(zap.String("type", e.Type()), zap.String("source", e.Source())).Info("Received event")

	eventType := e.Type()
	if eventType != event.EventTypeNotificationRequested.String() && eventType != event.EventTypeNotificationErrorRequested.String() {
		msg := "Unexpected event type"
		log.Error(msg, zap.String("received", e.Type()))
		return nil, fmt.Errorf("%s: %s", msg, e.Type())
	}

	// Parse internal notification data based on event type
	var (
		isErrorNotification bool
		requestID           string
		resultValue         float64
		errorInfo           *models.ErrorInfo
	)

	switch eventType {
	case event.EventTypeNotificationErrorRequested.String():
		isErrorNotification = true
		errorData := event.NotificationErrorRequestedData{}
		if err := e.DataAs(&errorData); err != nil {
			msg := "Failed to parse error event data"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		requestID = errorData.RequestID
		errorInfo = &models.ErrorInfo{
			Status:  errorData.Status,
			Code:    errorData.Code,
			Message: errorData.Message,
		}
		resultValue = -1.0
		log.With(zap.Any("notificationErrorData", errorData)).Error("Received notification error requested event")
	case event.EventTypeNotificationRequested.String():
		isErrorNotification = false
		data := event.NotificationRequestedData{}
		if err := e.DataAs(&data); err != nil {
			msg := "Failed to parse event data"
			log.With(zap.Error(err)).Error(msg)
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		log.With(zap.Any("notificationData", data)).Debug("Parsed notification requested data")
		requestID = data.RequestID
		resultValue = data.Result
	default:
		msg := "Unexpected event type"
		log.Error(msg, zap.String("received", e.Type()))
		return nil, fmt.Errorf("%s: %s", msg, e.Type())
	}

	// Atomically check and set notification flag FIRST to prevent any duplicates
	// This is the definitive check - only ONE handler instance will proceed past this point
	shouldSend, dbErr := h.db.TrySetNotificationSent(ctx, requestID)
	if dbErr != nil {
		log.With(zap.Error(dbErr)).Error("Failed to atomically check/set notificationSent flag")
		return nil, fmt.Errorf("failed to check notificationSent for job %s: %w", requestID, dbErr)
	}

	if !shouldSend {
		log.Info("Notification already sent by another instance, skipping duplicate")
		return nil, nil
	}

	log.Debug("Acquired exclusive right to send notification, proceeding")

	// Retrieve Job to determine sink and request kind
	job, err := h.db.GetJob(ctx, requestID)
	if err != nil {
		log.With(zap.Error(err)).Error("failed to read DB Job")
		return nil, fmt.Errorf("failed to read job %s from DB: %w", requestID, err)
	}
	log.With(zap.Any("job", job)).Debug("Fetched job for notification")

	// Check if subscription has expired
	if job.SubscriptionRequest.Config.SubscriptionExpireTime != nil {
		if time.Now().UTC().After(*job.SubscriptionRequest.Config.SubscriptionExpireTime) {
			log.With(
				zap.Time("expiredAt", *job.SubscriptionRequest.Config.SubscriptionExpireTime),
				zap.String("requestId", e.ID()),
			).Warn("Subscription has expired, notification not sent")
			// Return OK to stop knative retry loop
			return nil, nil
		}
	}

	sink := job.SubscriptionRequest.Sink
	if sink == "" {
		log.Error("job sink is empty; cannot deliver callback notification")
		return nil, fmt.Errorf("missing sink in subscription request for job %s", e.ID())
	}

	// Map internal RequestKind to CAMARA event type
	var camaraType models.EventTypeNotification
	switch job.RequestKind {
	case database.RequestKindEnergyConsumption:
		camaraType = models.EventTypeNotificationOrgCamaraprojectEnergyFootprintNotificationV1Energy
	case database.RequestKindCarbonFootprint:
		camaraType = models.EventTypeNotificationOrgCamaraprojectEnergyFootprintNotificationV1CarbonFootprint
	default:
		log.With(zap.String("requestKind", string(job.RequestKind))).Warn("unknown request kind; defaulting to energy event type")
		camaraType = models.EventTypeNotificationOrgCamaraprojectEnergyFootprintNotificationV1Energy
	}

	// Build CAMARA CloudEvent according to spec
	contentType := models.CloudEventDatacontenttype("application/json")
	dataMap := map[string]interface{}{
		"requestId": requestID,
	}

	switch job.RequestKind {
	case database.RequestKindEnergyConsumption:
		dataMap["energyConsumption"] = resultValue
	case database.RequestKindCarbonFootprint:
		dataMap["carbonFootprint"] = resultValue
	default:
		dataMap["energyConsumption"] = resultValue
	}

	cloudEvt := models.CloudEvent{
		Id:              e.ID(),
		Source:          models.Source(event.SourceEFNNotify.String()),
		Specversion:     models.N10,
		Type:            camaraType,
		Time:            models.DateTime(time.Now().UTC()),
		Datacontenttype: &contentType,
		Data:            &dataMap,
	}

	body, err := json.Marshal(cloudEvt)
	if err != nil {
		msg := "Failed to marshal cloud event for callback"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sink, bytes.NewReader(body))
	if err != nil {
		msg := "Failed to create HTTP request for callback"
		log.With(zap.Error(err)).Error(msg)
		return nil, fmt.Errorf("%s to sink %s: %w", msg, sink, err)
	}
	req.Header.Set("Content-Type", string(contentType))

	// Set x-correlator header if present in event extensions
	if xcorr, ok := e.Extensions()["x-correlator"]; ok {
		if xcorrStr, ok := xcorr.(string); ok && xcorrStr != "" {
			req.Header.Set("x-correlator", xcorrStr)
		}
	}

	// Set custom headers from protocolSettings if present
	if job.SubscriptionRequest.ProtocolSettings != nil && job.SubscriptionRequest.ProtocolSettings.Headers != nil {
		for key, value := range *job.SubscriptionRequest.ProtocolSettings.Headers {
			req.Header.Set(key, value)
			log.Debug("Set custom header from protocolSettings", zap.String("header", key))
		}
	}

	// Set Authorization header if sinkCredential is present
	if job.SubscriptionRequest.SinkCredential != nil {
		cred := job.SubscriptionRequest.SinkCredential
		switch cred.CredentialType {
		case models.SinkCredentialCredentialTypeACCESSTOKEN:
			at := cred.AccessTokenCredential
			if at.AccessToken != "" && at.AccessTokenType == models.AccessTokenCredentialAccessTokenTypeBearer {
				req.Header.Set("Authorization", "Bearer "+at.AccessToken)
			}
		default:
			log.Warn("credential type not implemented", zap.String("type", string(cred.CredentialType)))
		}
	}

	start := time.Now()
	client := h.getHTTPClient(sink)
	resp, err := client.Do(req)
	if err != nil {
		msg := "Failed to deliver notification to sink"
		log.With(zap.Error(err), zap.String("sink", sink)).Error(msg)
		return nil, fmt.Errorf("%s %s: %w", msg, sink, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := "Callback delivery returned non-success status"
		log.With(zap.Int("status", resp.StatusCode), zap.String("sink", sink)).Error(msg)
		return nil, fmt.Errorf("%s: sink %s returned status %d", msg, sink, resp.StatusCode)
	}

	logFields := []zap.Field{
		zap.String("sink", sink),
		zap.Int("status", resp.StatusCode),
		zap.Duration("latency", time.Since(start)),
		zap.String("camaraEventType", string(camaraType)),
		zap.Bool("isError", isErrorNotification),
		zap.Float64("result", resultValue),
	}

	if isErrorNotification {
		logFields = append(logFields, zap.Any("errorInfo", errorInfo))
	}

	log.With(logFields...).Info("Notification callback delivered successfully")

	// Emit internal event to indicate notification was sent
	return event.Event(requestID, event.EventTypeNotificationSent, event.SourceEFNNotify, nil)
}
