# Requirements

To build, deploy, and run the CAMARA Energy Footprint Notification (EFN) API, you need the following infrastructure and tools.

## Development Environment

*   **Go**: Version 1.24 or higher.
*   **Docker**: For building container images.
*   **Make** (Optional): For running build scripts if provided.
*   **OAPI-Codegen**: For regenerating API code from OpenAPI specs (`go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`).

## Infrastructure

The system is designed to run on Kubernetes with Knative Eventing.

### Kubernetes Cluster
*   A standard Kubernetes cluster (v1.24+).
*   Local options: [Kind](https://kind.sigs.k8s.io/), [Minikube](https://minikube.sigs.k8s.io/), or Docker Desktop.

### Knative Eventing & RabbitMQ
*   **Knative Eventing** must be installed on the cluster.
*   **RabbitMQ Broker**: The Helm chart is configured to use the `RabbitMQBroker` class. You must have the **Knative RabbitMQ Controller** installed.
*   **RabbitMQ Cluster**: A RabbitMQ cluster must be available in the Kubernetes cluster (e.g., via the RabbitMQ Cluster Operator). The Helm chart expects to link the broker to this cluster.

### Database
*   **MongoDB**: Version 5.0+.
*   The services require a connection string (URI) to a MongoDB instance.
*   For development, a simple MongoDB pod/service in the cluster is sufficient.
*   For production, a managed MongoDB service or a high-availability replica set is recommended.

### Policy Engine (Authorization)
*   **Cerbos**: The API uses Cerbos for policy-based authorization to control access to application instances.
*   Can be deployed using the Helm chart in `deploy/helm/cerbos`.
*   For development, authorization can be bypassed by setting `PDP_SKIP_POLICY_CHECK=true`.

### External Integrations

The EFN API relies on external systems for data retrieval:

*   **Cloud Observability Platform**: Provides energy consumption metrics for applications and network elements. The system includes configurable mock implementations for development.
*   **Traffic Volume Analytics**: Provides traffic volume measurements per network element. The system includes configurable mock implementations for development.
*   **Orchestrator**: Provides application instance metadata (IPs, associated network elements). Can be mocked using the dummy orchestrator.
*   For local development and testing, refer to the [MOCKED_COMPONENTS](MOCKED_COMPONENTS.md) documentation for details on using mock implementations of these external systems.
