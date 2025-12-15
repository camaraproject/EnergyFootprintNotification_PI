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
package models

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	OpenIdScopes = "openId.Scopes"
)

// Defines values for AMQPSettingsSenderSettlementMode.
const (
	Settled   AMQPSettingsSenderSettlementMode = "settled"
	Unsettled AMQPSettingsSenderSettlementMode = "unsettled"
)

// Defines values for AccessTokenCredentialAccessTokenType.
const (
	AccessTokenCredentialAccessTokenTypeBearer AccessTokenCredentialAccessTokenType = "bearer"
)

// Defines values for AccessTokenCredentialCredentialType.
const (
	AccessTokenCredentialCredentialTypeACCESSTOKEN  AccessTokenCredentialCredentialType = "ACCESSTOKEN"
	AccessTokenCredentialCredentialTypePLAIN        AccessTokenCredentialCredentialType = "PLAIN"
	AccessTokenCredentialCredentialTypeREFRESHTOKEN AccessTokenCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for CloudEventDatacontenttype.
const (
	Applicationjson CloudEventDatacontenttype = "application/json"
)

// Defines values for CloudEventSpecversion.
const (
	N10 CloudEventSpecversion = "1.0"
)

// Defines values for EventTypeNotification.
const (
	EventTypeNotificationOrgCamaraprojectEnergyFootprintNotificationV1CarbonFootprint EventTypeNotification = "org.camaraproject.energy-footprint-notification.v1.carbon-footprint"
	EventTypeNotificationOrgCamaraprojectEnergyFootprintNotificationV1Energy          EventTypeNotification = "org.camaraproject.energy-footprint-notification.v1.energy"
)

// Defines values for HTTPSettingsMethod.
const (
	POST HTTPSettingsMethod = "POST"
)

// Defines values for PlainCredentialCredentialType.
const (
	PlainCredentialCredentialTypeACCESSTOKEN  PlainCredentialCredentialType = "ACCESSTOKEN"
	PlainCredentialCredentialTypePLAIN        PlainCredentialCredentialType = "PLAIN"
	PlainCredentialCredentialTypeREFRESHTOKEN PlainCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for Protocol.
const (
	AMQP  Protocol = "AMQP"
	HTTP  Protocol = "HTTP"
	KAFKA Protocol = "KAFKA"
	MQTT3 Protocol = "MQTT3"
	MQTT5 Protocol = "MQTT5"
	NATS  Protocol = "NATS"
)

// Defines values for RefreshTokenCredentialAccessTokenType.
const (
	RefreshTokenCredentialAccessTokenTypeBearer RefreshTokenCredentialAccessTokenType = "bearer"
)

// Defines values for RefreshTokenCredentialCredentialType.
const (
	ACCESSTOKEN  RefreshTokenCredentialCredentialType = "ACCESSTOKEN"
	PLAIN        RefreshTokenCredentialCredentialType = "PLAIN"
	REFRESHTOKEN RefreshTokenCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for SubscriptionEventType.
const (
	SubscriptionEventTypeOrgCamaraprojectEnergyFootprintNotificationV1CarbonFootprint SubscriptionEventType = "org.camaraproject.energy-footprint-notification.v1.carbon-footprint"
	SubscriptionEventTypeOrgCamaraprojectEnergyFootprintNotificationV1Energy          SubscriptionEventType = "org.camaraproject.energy-footprint-notification.v1.energy"
)

// AMQPSettings defines model for AMQPSettings.
type AMQPSettings struct {
	Address              *string                           `json:"address,omitempty"`
	LinkName             *string                           `json:"linkName,omitempty"`
	LinkProperties       *map[string]string                `json:"linkProperties,omitempty"`
	SenderSettlementMode *AMQPSettingsSenderSettlementMode `json:"senderSettlementMode,omitempty"`
}

// AMQPSettingsSenderSettlementMode defines model for AMQPSettings.SenderSettlementMode.
type AMQPSettingsSenderSettlementMode string

// AMQPSubscriptionRequest defines model for AMQPSubscriptionRequest.
type AMQPSubscriptionRequest struct {
	// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
	// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
	// Specific event type attributes must be defined in `subscriptionDetail`
	// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
	Config Config `json:"config"`

	// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
	Protocol         Protocol      `json:"protocol"`
	ProtocolSettings *AMQPSettings `json:"protocolSettings,omitempty"`

	// Sink The address to which events shall be delivered using the selected protocol.
	Sink string `json:"sink"`

	// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
	SinkCredential *SinkCredential `json:"sinkCredential,omitempty"`

	// Types Camara Event types eligible to be delivered by this subscription.
	// Note: the maximum number of event types per subscription will be decided at API project level
	Types []SubscriptionEventType `json:"types"`
}

// AccessTokenCredential defines model for AccessTokenCredential.
type AccessTokenCredential struct {
	// AccessToken REQUIRED. An access token is a previously acquired token granting access to the target resource.
	AccessToken string `json:"accessToken"`

	// AccessTokenExpiresUtc REQUIRED. An absolute (UTC) timestamp at which the token shall be considered expired.
	// In the case of an ACCESS_TOKEN_EXPIRED termination reason, implementation should notify the client before the expiration date.
	// If the access token is a JWT and registered "exp" (Expiration Time) claim is present, the two expiry times should match.
	// It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	AccessTokenExpiresUtc time.Time `json:"accessTokenExpiresUtc"`

	// AccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
	AccessTokenType AccessTokenCredentialAccessTokenType `json:"accessTokenType"`

	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType AccessTokenCredentialCredentialType `json:"credentialType"`
}

// AccessTokenCredentialAccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
type AccessTokenCredentialAccessTokenType string

// AccessTokenCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type AccessTokenCredentialCredentialType string

// ApacheKafkaSettings defines model for ApacheKafkaSettings.
type ApacheKafkaSettings struct {
	AckMode               *int    `json:"ackMode,omitempty"`
	ClientId              *string `json:"clientId,omitempty"`
	PartitionKeyExtractor *string `json:"partitionKeyExtractor,omitempty"`
	TopicName             string  `json:"topicName"`
}

// ApacheKafkaSubscriptionRequest defines model for ApacheKafkaSubscriptionRequest.
type ApacheKafkaSubscriptionRequest struct {
	// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
	// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
	// Specific event type attributes must be defined in `subscriptionDetail`
	// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
	Config Config `json:"config"`

	// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
	Protocol         Protocol             `json:"protocol"`
	ProtocolSettings *ApacheKafkaSettings `json:"protocolSettings,omitempty"`

	// Sink The address to which events shall be delivered using the selected protocol.
	Sink string `json:"sink"`

	// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
	SinkCredential *SinkCredential `json:"sinkCredential,omitempty"`

	// Types Camara Event types eligible to be delivered by this subscription.
	// Note: the maximum number of event types per subscription will be decided at API project level
	Types []SubscriptionEventType `json:"types"`
}

// AppInstanceId A globally unique identifier associated with a running
// instance of an application.
// Edge Cloud Platform generates this identifier when the
// instantiation in the Edge Cloud Zone is successful
type AppInstanceId = openapi_types.UUID

// CloudEvent The notification callback
type CloudEvent struct {
	// Data Event details payload described in each CAMARA API and referenced by its type
	Data *map[string]interface{} `json:"data,omitempty"`

	// Datacontenttype media-type that describes the event payload encoding, must be "application/json" for CAMARA APIs
	Datacontenttype *CloudEventDatacontenttype `json:"datacontenttype,omitempty"`

	// Id identifier of this event, that must be unique in the source context.
	Id string `json:"id"`

	// Source Identifies the context in which an event happened - be a non-empty
	// `URI-reference` like:
	// - URI with a DNS authority:
	//   * https://github.com/cloudevents
	//   * mailto:cncf-wg-serverless@lists.cncf.io
	// - Universally-unique URN with a UUID:
	//   * urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66
	// - Application-specific identifier:
	//   * /cloudevents/spec/pull/123
	//   * 1-555-123-4567
	Source Source `json:"source"`

	// Specversion Version of the specification to which this event conforms (must be 1.0 if it conforms to cloudevents 1.0.2 version)
	Specversion CloudEventSpecversion `json:"specversion"`

	// Time Timestamp of when the occurrence happened. Must adhere to RFC 3339.
	// WARN: This optional field in CloudEvents specification is required in
	// CAMARA APIs implementation.
	Time DateTime `json:"time"`

	// Type Event triggered when an event-type event occurred.
	Type EventTypeNotification `json:"type"`
}

// CloudEventDatacontenttype media-type that describes the event payload encoding, must be "application/json" for CAMARA APIs
type CloudEventDatacontenttype string

// CloudEventSpecversion Version of the specification to which this event conforms (must be 1.0 if it conforms to cloudevents 1.0.2 version)
type CloudEventSpecversion string

// CloudEventCarbonFootprint The notification callback
type CloudEventCarbonFootprint = CloudEvent

// CloudEventEnergy The notification callback
type CloudEventEnergy = CloudEvent

// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
// Specific event type attributes must be defined in `subscriptionDetail`
// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
type Config struct {
	// InitialEvent Set to `true` by API consumer if consumer wants to get an event as soon as the subscription is created and current situation reflects event request.
	// Example: Consumer request Roaming event. If consumer sets initialEvent to true and device is in roaming situation, an event is triggered
	// Up to API project decision to keep it.
	InitialEvent *bool `json:"initialEvent,omitempty"`

	// SubscriptionDetail The detail of the requested event subscription.
	SubscriptionDetail CreateSubscriptionDetail `json:"subscriptionDetail"`

	// SubscriptionExpireTime The subscription expiration time (in date-time format) requested by the API consumer. It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone. Up to API project decision to keep it.
	SubscriptionExpireTime *time.Time `json:"subscriptionExpireTime,omitempty"`

	// SubscriptionMaxEvents Identifies the maximum number of event reports to be generated (>=1) requested by the API consumer - Once this number is reached, the subscription ends. Up to API project decision to keep it.
	SubscriptionMaxEvents *int `json:"subscriptionMaxEvents,omitempty"`
}

// CreateSubscriptionDetail The detail of the requested event subscription.
type CreateSubscriptionDetail = map[string]interface{}

// DateTime Timestamp of when the occurrence happened. Must adhere to RFC 3339.
// WARN: This optional field in CloudEvents specification is required in
// CAMARA APIs implementation.
type DateTime = time.Time

// ErrorInfo defines model for ErrorInfo.
type ErrorInfo struct {
	// Code A human-readable code to describe the error
	Code string `json:"code"`

	// Message A human-readable description of what the event represents
	Message string `json:"message"`

	// Status HTTP response status code
	Status int `json:"status"`
}

// EventTypeNotification Event triggered when an event-type event occurred.
type EventTypeNotification string

// HTTPSettings defines model for HTTPSettings.
type HTTPSettings struct {
	// Headers A set of key/value pairs that is copied into the HTTP request as custom headers.
	//
	// NOTE: Use/Applicability of this concept has not been discussed in Commonalities under the scope of Meta Release v0.4. When required by an API project as an option to meet a UC/Requirement, please generate an issue for Commonalities discussion about it.
	Headers *map[string]string `json:"headers,omitempty"`

	// Method The HTTP method to use for sending the message.
	Method *HTTPSettingsMethod `json:"method,omitempty"`
}

// HTTPSettingsMethod The HTTP method to use for sending the message.
type HTTPSettingsMethod string

// MQTTSettings defines model for MQTTSettings.
type MQTTSettings struct {
	Expiry         *int32                  `json:"expiry,omitempty"`
	Qos            *int32                  `json:"qos,omitempty"`
	Retain         *bool                   `json:"retain,omitempty"`
	TopicName      string                  `json:"topicName"`
	UserProperties *map[string]interface{} `json:"userProperties,omitempty"`
}

// MQTTSubscriptionRequest defines model for MQTTSubscriptionRequest.
type MQTTSubscriptionRequest struct {
	// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
	// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
	// Specific event type attributes must be defined in `subscriptionDetail`
	// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
	Config Config `json:"config"`

	// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
	Protocol         Protocol      `json:"protocol"`
	ProtocolSettings *MQTTSettings `json:"protocolSettings,omitempty"`

	// Sink The address to which events shall be delivered using the selected protocol.
	Sink string `json:"sink"`

	// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
	SinkCredential *SinkCredential `json:"sinkCredential,omitempty"`

	// Types Camara Event types eligible to be delivered by this subscription.
	// Note: the maximum number of event types per subscription will be decided at API project level
	Types []SubscriptionEventType `json:"types"`
}

// NATSSettings defines model for NATSSettings.
type NATSSettings struct {
	Subject string `json:"subject"`
}

// NATSSubscriptionRequest defines model for NATSSubscriptionRequest.
type NATSSubscriptionRequest struct {
	// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
	// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
	// Specific event type attributes must be defined in `subscriptionDetail`
	// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
	Config Config `json:"config"`

	// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
	Protocol         Protocol      `json:"protocol"`
	ProtocolSettings *NATSSettings `json:"protocolSettings,omitempty"`

	// Sink The address to which events shall be delivered using the selected protocol.
	Sink string `json:"sink"`

	// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
	SinkCredential *SinkCredential `json:"sinkCredential,omitempty"`

	// Types Camara Event types eligible to be delivered by this subscription.
	// Note: the maximum number of event types per subscription will be decided at API project level
	Types []SubscriptionEventType `json:"types"`
}

// PlainCredential defines model for PlainCredential.
type PlainCredential struct {
	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType PlainCredentialCredentialType `json:"credentialType"`

	// Identifier The identifier might be an account or username.
	Identifier string `json:"identifier"`

	// Secret The secret might be a password or passphrase.
	Secret string `json:"secret"`
}

// PlainCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type PlainCredentialCredentialType string

// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
type Protocol string

// RefreshTokenCredential defines model for RefreshTokenCredential.
type RefreshTokenCredential struct {
	// AccessToken REQUIRED. An access token is a previously acquired token granting access to the target resource.
	AccessToken *string `json:"accessToken,omitempty"`

	// AccessTokenExpiresUtc REQUIRED. An absolute (UTC) timestamp at which the token shall be considered expired.
	// In the case of an ACCESS_TOKEN_EXPIRED termination reason, implementation should notify the client before the expiration date.
	// If the access token is a JWT and registered "exp" (Expiration Time) claim is present, the two expiry times should match.
	// It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	AccessTokenExpiresUtc *time.Time `json:"accessTokenExpiresUtc,omitempty"`

	// AccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
	AccessTokenType *RefreshTokenCredentialAccessTokenType `json:"accessTokenType,omitempty"`

	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType RefreshTokenCredentialCredentialType `json:"credentialType"`

	// RefreshToken REQUIRED. An refresh token credential used to acquire access tokens.
	RefreshToken *string `json:"refreshToken,omitempty"`

	// RefreshTokenEndpoint REQUIRED. A URL at which the refresh token can be traded for an access token.
	RefreshTokenEndpoint *string `json:"refreshTokenEndpoint,omitempty"`
}

// RefreshTokenCredentialAccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
type RefreshTokenCredentialAccessTokenType string

// RefreshTokenCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type RefreshTokenCredentialCredentialType string

// ReportCreationRequest resource containing the service under analysis and the callback information for the API Consumer to be notified with the results of the analysis. If no "timePeriod" is provided the analysis is performed from the activation of the first instance of the Application.
type ReportCreationRequest struct {
	// RequestId Identifier for the request. This parameter is returned by the API and must be used to update it.
	RequestId *string `json:"requestId,omitempty"`

	// Service list of Application Instance Identifiers. This are the instances of the applications producing the service under analysis.
	Service []AppInstanceId `json:"service"`

	// SubscriptionRequest The request for creating a event-type event subscription
	SubscriptionRequest SubscriptionRequest `json:"subscriptionRequest"`
	TimePeriod          *TimePeriod         `json:"timePeriod,omitempty"`
}

// Source Identifies the context in which an event happened - be a non-empty
// `URI-reference` like:
// - URI with a DNS authority:
//   - https://github.com/cloudevents
//   - mailto:cncf-wg-serverless@lists.cncf.io
//
// - Universally-unique URN with a UUID:
//   - urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66
//
// - Application-specific identifier:
//   - /cloudevents/spec/pull/123
//   - 1-555-123-4567
type Source = string

// SubscriptionEventType event-type that could be subscribed through this subscription. Several event-type could be defined.
type SubscriptionEventType string

// TimePeriod defines model for TimePeriod.
type TimePeriod struct {
	// EndDate An instant of time, ending of the TimePeriod. If not included, then the period has no ending date. It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	EndDate *time.Time `json:"endDate,omitempty"`

	// StartDate An instant of time, starting of the TimePeriod. It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	StartDate time.Time `json:"startDate"`
}

// XCorrelator defines model for XCorrelator.
type XCorrelator = string

// Generic400 defines model for Generic400.
type Generic400 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic401 defines model for Generic401.
type Generic401 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic403 defines model for Generic403.
type Generic403 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic404 defines model for Generic404.
type Generic404 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic409 defines model for Generic409.
type Generic409 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic410 defines model for Generic410.
type Generic410 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic422 defines model for Generic422.
type Generic422 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic429 defines model for Generic429.
type Generic429 struct {
	Code interface{} `json:"code"`

	// Message A human-readable description of what the event represents
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// CalculateCarbonFootprintParams defines parameters for CalculateCarbonFootprint.
type CalculateCarbonFootprintParams struct {
	// XCorrelator Correlation id for the different services
	XCorrelator *XCorrelator `json:"x-correlator,omitempty"`
}

// CalculateEnergyConsumptionParams defines parameters for CalculateEnergyConsumption.
type CalculateEnergyConsumptionParams struct {
	// XCorrelator Correlation id for the different services
	XCorrelator *XCorrelator `json:"x-correlator,omitempty"`
}

// CalculateCarbonFootprintJSONRequestBody defines body for CalculateCarbonFootprint for application/json ContentType.
type CalculateCarbonFootprintJSONRequestBody = ReportCreationRequest

// CalculateEnergyConsumptionJSONRequestBody defines body for CalculateEnergyConsumption for application/json ContentType.
type CalculateEnergyConsumptionJSONRequestBody = ReportCreationRequest
