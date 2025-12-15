# Architecture

This project implements the CAMARA Energy Footprint Notification (EFN) API using an event-driven microservices architecture.

## Overview

The system is composed of several decoupled services that communicate primarily through CloudEvents and a shared MongoDB database. The EFN API enables API consumers to retrieve energy consumption and carbon footprint information for services deployed on the operator's infrastructure.

### Components

1.  **API Service (`cmd/api`)**
    *   **Role**: Entry point for API consumers.
    *   **Responsibilities**:
        *   Validates incoming requests against the OpenAPI specification.
        *   Authorizes access to requested application instances via Policy Decision Point (PDP).
        *   Creates job records in MongoDB.
        *   Publishes `gatherinfo.requested` events to the event broker.
    *   **Endpoints**:
        *   `POST /calculate-energy-consumption` - Calculate energy consumption for specified applications.
        *   `POST /calculate-carbon-footprint` - Calculate carbon footprint for specified applications.

2.  **Worker Service (`cmd/worker`)**
    *   **Role**: Performs the actual energy/carbon calculations.
    *   **Responsibilities**:
        *   Listens for multiple event types (`gatherinfo.requested`, `app.consumption.requested`, `networkelement.energy.requested`, `networkelement.traffic.requested`, `calculation.requested`).
        *   For each application instance:
            *   Retrieves energy consumption from the Cloud Observability interface.
            *   Retrieves traffic volumes from the Traffic Volume interface.
            *   Calculates network elements energy consumption proportionally.
        *   Aggregates results and calculates total energy consumption or carbon footprint.
        *   Stores calculation results to  MongoDB.
        *   Publishes `notification.requested` or `notification.error.requested` events.

3.  **Notification Service (`cmd/notification`)**
    *   **Role**: Handles callbacks to the API consumer.
    *   **Responsibilities**:
        *   Listens for `notification.requested` and `notification.error.requested` events.
        *   Retrieves the full job result from MongoDB.
        *   Sends a webhook notification to the `sink` URL provided in the initial subscription request.
        *   Publishes `notification.sent` event after successful delivery.

4.  **Sink Receiver (`cmd/sinkreceiver`)**
    *   **Role**: Testing utility.
    *   **Responsibilities**:
        *   Acts as a mock endpoint for receiving webhook notifications during local development.
        *   Logs all incoming requests for debugging purposes.

## Knative Eventing

The system relies on Knative Eventing for asynchronous communication. The following table describes the events and their flow through the system.

| Event Type | Source | Producer | Consumer(s) | Description |
| :--- | :--- | :--- | :--- | :--- |
| `it.tim.efn.gatherinfo.requested` | `urn:tim:efn-api` | **API** | **Worker** | Sent when a user requests energy consumption or carbon footprint calculation. Contains the job ID and request details. |
| `it.tim.efn.app.consumption.requested` | `urn:tim:efn-worker` | **Worker** | **Worker** | Sent to get energy consumption for an application instance. |
| `it.tim.efn.networkelement.energy.requested` | `urn:tim:efn-worker` | **Worker** | **Worker** | Sent to get energy consumption for a single network element. |
| `it.tim.efn.networkelement.traffic.requested` | `urn:tim:efn-worker` | **Worker** | **Worker** | Sent to get traffic volume info for network elements. |
| `it.tim.efn.calculation.requested` | `urn:tim:efn-worker` | **Worker** | **Worker** | Sent to trigger final calculation after all values are gathered. |
| `it.tim.efn.notification.requested` | `urn:tim:efn-worker` | **Worker** | **Notification** | Sent when the calculation has completed successfully. |
| `it.tim.efn.notification.error.requested` | `urn:tim:efn-worker` | **Worker** | **Notification** | Sent when an error occurs during processing. |
| `it.tim.efn.notification.sent` | `urn:tim:efn-notification` | **Notification** | N/A | Sent when a notification has been delivered. |

### Triggers

The following Knative Triggers are defined to route events from the Broker to the services:

*   `gatherinfo-requested-trigger`: Routes `gatherinfo.requested` -> `efn-worker`.
*   `app-consumption-requested-trigger`: Routes `app.consumption.requested` -> `efn-worker`.
*   `networkelement-energy-requested-trigger`: Routes `networkelement.energy.requested` -> `efn-worker`.
*   `networkelement-traffic-requested-trigger`: Routes `networkelement.traffic.requested` -> `efn-worker`.
*   `calculation-requested-trigger`: Routes `calculation.requested` -> `efn-worker`.
*   `notification-requested-trigger`: Routes `notification.requested` -> `efn-notification`.
*   `notification-error-requested-trigger`: Routes `notification.error.requested` -> `efn-notification`.

## Data Flow

1.  **Request**: User sends `POST /calculate-energy-consumption` or `POST /calculate-carbon-footprint`.
2.  **Validation**: API validates request against OpenAPI spec and time period constraints.
3.  **Authorization**: API checks if user is authorized to access the requested application instances.
4.  **Persistence**: API creates a Job with the request information in MongoDB.
5.  **Event**: API sends `gatherinfo.requested` to Broker.
6.  **Processing**: Worker receives event and for each application:
    *   Retrieves application energy consumption from Cloud Observability.
    *   Retrieves traffic volumes for network elements.
    *   Calculates proportional network energy consumption.
7.  **Calculation**: Worker aggregates all results and calculates final energy/carbon value.
8.  **Completion**: Worker updates job status to `completed` and sends `notification.requested`.
9.  **Notification**: Notification service receives completion event and sends webhook to user's sink URL, then emits `notification.sent`.

