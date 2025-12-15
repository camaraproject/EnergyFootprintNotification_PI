# Documentation Overview

This directory contains the documentation for the **Energy Footprint Notification (EFN) API** Provider Implementation, part of the CAMARA project.

The EFN API enables API consumers (e.g., Service Providers) to retrieve information about energy consumption and carbon footprint for services running on the operator's infrastructure. The system uses an event-driven microservices architecture with Knative Eventing and MongoDB.

## Documentation Index

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture, components, event flows, and database schema |
| [CONFIGURATION.md](CONFIGURATION.md) | Environment variables and Helm configuration reference |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Step-by-step deployment guide for Kubernetes with Helm |
| [REQUIREMENTS.md](REQUIREMENTS.md) | Infrastructure and development environment requirements |
| [MOCKED_COMPONENTS.md](MOCKED_COMPONENTS.md) | Available mocking mechanisms for development and testing |

## Quick Links

- **API Specification**: [`api/energy-footprint-notification.yaml`](../api/energy-footprint-notification.yaml)
- **CAMARA Project**: [EnergyFootprintNotification](https://github.com/camaraproject/EnergyFootprintNotification)
- **Main README**: [`../README.md`](../README.md)
