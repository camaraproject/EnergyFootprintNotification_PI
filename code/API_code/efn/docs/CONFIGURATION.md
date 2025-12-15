# Configuration Reference

The application is configured via environment variables, which are mapped to Helm values.

## Environment Variables

### Logging
| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Logging format (`json`, `console`) | `production` (json) |

### API Service
| Variable | Description | Default |
|----------|-------------|---------|
| `API_ADDRESS` | HTTP listen address | `0.0.0.0:8080` |
| `API_MAX_TIME_PERIOD_DAYS` | Maximum allowed time period in days for historical data queries | `730` (2 years) |
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `efn` |
| `PDP_ADDRESS` | Cerbos policy engine address | `http://localhost:3593` |
| `PDP_SKIP_POLICY_CHECK` | Bypass authorization (DEV ONLY) | `false` |

### Worker Service
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `efn` |
| `K_SINK` | CloudEvents sink URL (set by Knative SinkBinding) | - |
| `CLIENT_TYPE` | Cloud Observability client type (`configurable`, `dummy`, or default) | `dummy` |
| `TRAFFIC_CLIENT_TYPE` | Traffic Volume client type (`configurable` or `dummy`) | `dummy` |
| `CARBON_FACTOR_TCO2E_PER_KWH` | CO2 conversion factor (tCO2e per kWh) | `0.00035` |

#### Cloud Observability Configurable Client
| Variable | Description | Default |
|----------|-------------|---------|
| `CLOUDOBS_CONFIG_APP_VALUE` | Default app energy consumption value (kWh) | `0.0020` |
| `CLOUDOBS_CONFIG_NE_VALUE` | Default NE energy consumption value (kWh) | `0.0010` |
| `CLOUDOBS_CONFIG_SUCCESS_COUNT` | If >0, always succeed for N requests | `0` |
| `CLOUDOBS_CONFIG_ERROR_COUNT` | If >0 and success_count=0, always fail for N requests | `0` |
| `CLOUDOBS_CONFIG_ERROR_TYPE` | Error type: `throttling` or `permanent` | `throttling` |
| `CLOUDOBS_CONFIG_DELAY_MS` | Request processing delay in milliseconds | `0` |

#### Cloud Observability Error Dummy (for DLQ testing)
| Variable | Description | Default |
|----------|-------------|---------|
| `CLOUDOBS_FAIL_THROTTLE` | If `true`, simulate throttling errors | `false` |
| `CLOUDOBS_FAIL_NE` | If `true`, simulate network element errors | `false` |

#### Traffic Volume Configurable Client
| Variable | Description | Default |
|----------|-------------|---------|
| `TRAFFIC_CONFIG_IP_VOLUME` | Default app instance IP traffic volume (Mbps) | `100.0` |
| `TRAFFIC_CONFIG_ALL_VOLUME` | Default total NE traffic volume (Mbps) | `1000.0` |
| `TRAFFIC_CONFIG_SUCCESS_COUNT` | If >0, always succeed for N requests | `0` |
| `TRAFFIC_CONFIG_ERROR_COUNT` | If >0 and success_count=0, always fail for N requests | `0` |
| `TRAFFIC_CONFIG_ERROR_TYPE` | Error type: `throttling` or `permanent` | `throttling` |
| `TRAFFIC_CONFIG_DELAY_MS` | Request processing delay in milliseconds | `0` |

### Notification Service
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `efn` |
| `K_SINK` | CloudEvents sink URL (set by Knative SinkBinding) | - |
| `HTTP_INSECURE_SKIP_VERIFY` | Skip TLS verification for internal services | `false` |

### Sink Receiver Service (Testing Only)
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP listen port | `8090` |
| `EXPECTED_RESULT_VALUE` | Expected result value for validation | `0.0044` |

## Helm Values (`values.yaml`)

This are all the configurable Helm values with their default settings:
```yaml
image:
  repository: ghcr.io/your_repository/camara-efn-api
  tag: "0.0.3"
  pullPolicy: Always

imagePullSecrets: []

knative:
  broker:
    name: rabbitmq-broker
    class: RabbitMQBroker
    delivery:
      retry: 3
      backoffPolicy: exponential
      backoffDelay: PT1S
      deadLetterPath: /dlq
  namespace: demo
  triggers:
    gatherInfoRequested:
      parallelism: 100
    appConsumptionRequested:
      parallelism: 100
    networkElementEnergyRequested:
      parallelism: 100
    networkElementTrafficRequested:
      parallelism: 100
    calculationRequested:
      parallelism: 100
    notificationRequested:
      parallelism: 100
  services:
    api:
      minScale: 1
      maxScale: 3
      target: 1
      metric: concurrency
    worker:
      minScale: 1
      maxScale: 16
      target: 1
      metric: concurrency
      scaleDownDelay: 30s
    notification:
      minScale: 1
      maxScale: 3
      target: 1
      metric: concurrency
  rabbitmq:
    brokerConfig:
      name: rabbitmq-broker-config
      queueType: quorum
      cluster:
        name: my-rabbit
        namespace: ""

api:
  maxTimePeriodDays: 730

logger:
  level: debug
  format: development

pdp:
  address: "cerbos.demo.svc.cluster.local:3593"
  skip: false

database:
  name: efn
  uri: "mongodb://root:CHANGEME@mongo-mongodb.data.svc.cluster.local:27017"

http:
  insecureSkipVerify: false

cloudObservability:
  failThrottle: false
  failNE: false

calculator:
  carbonFactorTCO2ePerKWh: 0.00035

mock:
  enabled: false
  cloudObservability:
    appValue: "0.0020"
    neValue: "0.0010"
    successCount: "0"
    errorCount: "0"
    errorType: "throttling"
    delayMS: "0"
  trafficVolume:
    ipVolume: "100.0"
    allVolume: "1000.0"
    successCount: "0"
    errorCount: "0"
    errorType: "throttling"
    delayMS: "0"
```
