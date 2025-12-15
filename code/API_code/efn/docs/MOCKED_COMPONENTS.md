# Mocked Components

To facilitate development and testing without requiring access to live infrastructure observability systems or network analytics, this project includes several mocking mechanisms.

## 1. Configurable Cloud Observability Client

The **Worker** service includes a "Configurable" implementation of the `CloudObservability` interface for development and testing.

*   **Interface**: `pkg/cloudobservability/interface.go`
*   **Implementation**: `pkg/cloudobservability/configurable.go`
*   **Behavior**:
    *   **RetrieveAppEnergyConsumption**: Returns configurable mock energy consumption values for application instances.
    *   **RetrieveNetworkElementEnergyConsumption**: Returns configurable mock energy consumption values for network elements.
*   **Use Case**: Unit testing, local development where no real observability platform is available.

## 2. Configurable Traffic Volume Client

The **Worker** service also includes a "Configurable" implementation of the `TrafficVolume` interface.

*   **Interface**: `pkg/trafficvolume/interface.go`
*   **Implementation**: `pkg/trafficvolume/configurable.go`
*   **Behavior**:
    *   **RetrieveTrafficVolumes**: Returns mock traffic volume measurements for network elements, including both per-IP and total NE volumes.
*   **Use Case**: Testing the proportional energy calculation logic without a real traffic analytics system.

## 3. Dummy Orchestrator Client

The system includes a dummy implementation of the `Orchestrator` interface.

*   **Interface**: `pkg/orchestrator/interface.go`
*   **Implementation**: `pkg/orchestrator/dummy.go`
*   **Role**: Returns static application topology and network element information that provides application instance metadata.
*   **Behavior**: Returns static or configurable application instance information (IPs, network elements, infrastructure types).

## 4. Sink Receiver (`cmd/sinkreceiver`)

The **Sink Receiver** is a standalone service for testing webhook notifications.

*   **Role**: Acts as a mock callback endpoint during local development.
*   **Usage**:
    1.  Deploy the `sinkreceiver` (chart available in `deploy/helm/sinkreceiver`).
    2.  Use its URL as the `sink` in your subscription requests.
*   **Behavior**: Accepts any incoming HTTP request and logs the body/headers. This allows you to verify that the Notification service is sending the correct callbacks.

## 5. Policy Bypass (AllowAll)

The API service supports bypassing the Cerbos policy engine for development.

*   **Implementation**: `pkg/policy/allowall.go`
*   **Activation**: Set `PDP_SKIP_POLICY_CHECK=true` environment variable (or `pdp.skipPolicyCheck: true` in Helm).
*   **Behavior**: All authorization checks return success, allowing any API consumer to access any application instance.
*   **Warning**: This should **never** be used in production environments.

## 6. Basic Calculator

The system includes a basic implementation of the energy/carbon calculator.

*   **Interface**: `pkg/calculator/interface.go`
*   **Implementation**: `pkg/calculator/basic.go`
*   **Behavior**:
    *   **CalculateEnergyConsumption**: Aggregates energy consumption from application and network element data.
    *   **CalculateCarbonFootprint**: Converts energy consumption to CO2 equivalent using a configurable conversion factor.
*   **Use Case**: Provides a reference implementation that can be replaced with more sophisticated calculation models.
