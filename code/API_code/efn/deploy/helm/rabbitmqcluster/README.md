# rabbitmqcluster Helm Chart

Deploys a `RabbitmqCluster` CR (via RabbitMQ Cluster Operator) and optional application queues used alongside the Knative RabbitMQ Broker defined in the `api` chart.

## Contents
- RabbitmqCluster (name configurable)
- Optional list of Queue CRs (NOT the internal Broker/Trigger queues; those are created automatically by eventing-rabbitmq)

## When to use
Use this chart when you want to provision and manage the RabbitMQ cluster lifecycle separately from the API/eventing chart, keeping concerns isolated.

## Values Overview
```yaml
rabbitmq:
  cluster:
    name: my-rabbit          # Cluster name referenced by BrokerConfig
    replicas: 1
    namespaceOverride: ""    # Uses release namespace if empty
    additionalConfig: |      # Extra rabbitmq.conf style lines
      # ERL_MAX_PORTS=4096
    defaultVhost:            # Optional default vhost
  persistence:
    enabled: true
    storageClass: ""
    size: 10Gi
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 1Gi
queues:
  - name: demo-queue
    vhost: "/"
    durable: true
    autoDelete: false
    type: quorum
    arguments: {}
createQueues: true
```

## Installation
```bash
helm upgrade --install rabbitmq ./deploy/helm/rabbitmqcluster -n demo
```

## Verifying
```bash
kubectl get rabbitmqcluster my-rabbit -n demo
kubectl get queues -n demo
```

## Integration with API Chart
Set in `deploy/helm/api/values.yaml`:
```yaml
knative:
  rabbitmq:
    brokerConfig:
      cluster:
        name: my-rabbit
```
Same namespace must match where this chart deploys the cluster (demo by default).

## Notes
- Internal Knative Broker/Trigger queues are created by the eventing-rabbitmq controllers; do not list them here.
- Quorum queues are recommended for production durability; use classic for lightweight dev setups.
- Add entries under `arguments` for advanced features (e.g. `x-message-ttl`).

## Troubleshooting
If Knative Triggers show condition `ExchangeCredentialsUnavailable`:
1. Check the cluster conditions:
  ```bash
  kubectl get rabbitmqcluster my-rabbit -n demo -o yaml | grep -A5 conditions:
  ```
2. Verify the default user secret exists:
  ```bash
  kubectl get secret -n demo | grep my-rabbit
  kubectl describe secret my-rabbit-default-user -n demo || true
  ```
  Required keys: `username`, `password` (optional: `port`, `uri`).
3. Ensure BrokerConfig points to the correct cluster name/namespace.
4. Recreate Broker after cluster becomes Ready:
  ```bash
  kubectl delete broker rabbitmq-broker -n demo
  helm upgrade --install efn-api ./deploy/helm/api -n demo
  ```
5. Describe the cluster for reconciliation errors:
  ```bash
  kubectl describe rabbitmqcluster my-rabbit -n demo
  ```
6. If secret missing: the cluster may not have finished reconciling—check operator pods logs.

This chart uses `override.statefulSet` to set container resources; previously using unsupported top-level `resources` can prevent secret creation and lead to `ExchangeCredentialsUnavailable`.

## Future Enhancements
- TLS configuration support
- Policies CR management
- Automatic dead-letter queues
