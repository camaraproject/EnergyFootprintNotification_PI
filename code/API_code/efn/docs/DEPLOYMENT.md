# Deployment Guide

This guide explains how to deploy the CAMARA Energy Footprint Notification (EFN) API to a Kubernetes cluster using Helm.

## Prerequisites

1.  A Kubernetes cluster with **Knative Eventing** installed.
2.  **RabbitMQ Cluster** and **Knative RabbitMQ Controller** (the Helm chart uses `RabbitMQBroker`).
3.  **Helm** 3.0+ installed locally.
4.  **MongoDB** running in the cluster (or accessible from it).
5.  The system is made to integrate with external components: Policy Decision Point, Orchestrator, Cloud Observability and Traffic Volume. Mocks for these components can be found in [MOCKED_COMPONENTS](MOCKED_COMPONENTS.md)

## Build Images

Before deploying, you need to build the Docker images for the microservices.

```bash
# Build API
docker build -f build/package/docker/api.Dockerfile -t your-registry/efn-api:latest .

# Build Worker
docker build -f build/package/docker/worker.Dockerfile -t your-registry/efn-worker:latest .

# Build Notification
docker build -f build/package/docker/notification.Dockerfile -t your-registry/efn-notification:latest .

# Push images
docker push your-registry/efn-api:latest
# ... repeat for others
```

## Helm Deployment

The project includes a Helm chart in `deploy/helm/api`.

### 1. Configure Values

Create a `my-values.yaml` file to override the defaults.

```yaml
# my-values.yaml

# Image configuration
image:
  repository: your-registry/camara-efn-api
  tag: "0.0.3"
  pullPolicy: Always

# Knative namespace where services will be deployed
knative:
  namespace: camara-efn
  broker:
    name: rabbitmq-broker
  rabbitmq:
    brokerConfig:
      cluster:
        name: my-rabbit  # Name of your RabbitMQ cluster

# Database Connection
database:
  uri: "mongodb://username:password@mongo-service:27017"
  name: "efn"

# API Configuration
api:
  maxTimePeriodDays: 730

# Logging
logger:
  level: info
  format: production

# Cerbos Policy Engine
pdp:
  address: "cerbos.camara-efn.svc.cluster.local:3593"
  skip: false  # Set to true for development without auth

# Calculator
calculator:
  carbonFactorTCO2ePerKWh: 0.00035

# Mock configuration (for development/testing)
mock:
  enabled: false  # Set to true to use configurable mock clients
```

### 2. Install Chart

```bash
helm install efn-api ./deploy/helm/api \
  --namespace camara-efn \
  --create-namespace \
  -f my-values.yaml
```

### 3. Verify Deployment

Check that all pods are running:

```bash
kubectl get pods -n camara-efn
```

You should see pods for:
*   `efn-api`
*   `efn-worker`
*   `efn-notification`

Check that the Knative triggers are ready:

```bash
kubectl get triggers -n camara-efn
```

## Local Development (Kind/Minikube)

For local testing, you can use the provided `sinkreceiver` to mock callback endpoints.

1.  **Install MongoDB**:
    ```bash
    helm install mongo oci://registry-1.docker.io/bitnamicharts/mongodb -n camara-efn
    ```

2.  **Deploy the API with Mocked Components**:
    The system includes configurable mock implementations for Cloud Observability and Traffic Volume interfaces.

    ```bash
    helm install efn-api ./deploy/helm/api \
      --namespace camara-efn \
      --set knative.namespace=camara-efn \
      --set database.uri="mongodb://root:password@mongo-mongodb.camara-efn.svc.cluster.local:27017" \
      --set pdp.skip=true \
      --set mock.enabled=true
    ```

3.  **Deploy the Sink Receiver** (optional, for testing notifications):
    ```bash
    helm install sinkreceiver ./deploy/helm/sinkreceiver --namespace camara-efn
    ```

## Testing

The repository includes end-to-end test scripts in the `test/e2e` directory:

```bash
# Single request test
./test/e2e/test-energy-consumption.sh

# Parallel requests test (100 app instances)
./test/e2e/test-energy-consumption-app-instances-100.sh

# Stress test (200 app instances)
./test/e2e/test-energy-consumption-app-instances-200.sh
```
