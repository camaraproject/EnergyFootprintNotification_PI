# Energy Footprint Notification API - Provider Implementation

This repository contains the Provider Implementatin for the Energy Footprint Notification API. This is a possible implementation of the Transformation Function for the Energy Footprint Notification API.

## Energy Footprint Notification API - Description

This document outlines the implementation details for the Energy Footprint Notification API.

The specificatin of the API is here: https://github.com/camaraproject/EnergyFootprintNotification

The Transformation Function performs the following tasks:

1) verifies if the API Consumer is allowed to invoke the API for the specific requested applications
2) gathers all the information related to the specific applications. Information that are required to calculate the energy consumption and the carbon footprint in the specified time period. For each requested application:
  - The public IP address exposed by the application
  - The identifiers for the instances of the Network Elements (e.g. UPF) supporting the service (traversed by the traffic flow to and from the application)
  - The vendor of the Network Elements
  - The infrastructure type supporting the Network Elements 
  - The infrastructure type supporting the application
3) calculates the percentage of the traffic flow in each Network Element that is related to the service under observation
4) calculates the overall energy consumption of the involved Network Elements
5) Makes a proportion to estimate the energy consumption for the service
6) Calculates the energy consumption for the applications
7) Estimates the E2E energy consumption 

### Implemented Version

r1.3

### API Functionality

The Energy Footprint Notification API supports the following intents:

  - Intent1: Which is the overall energy consumption for my service in this period of time?
  - Intent2: Which is the overall carbon footprint for my service in this period of time?

To support the above intents two enpoints are provided:

  - **calculate-energy-consumption:** Provides to the API Consumer the requested information about the energy
    consumption for the target service, considering all the active instances, in the period of time of interest
  - **calculate-energy-consumption:** Provides to the API Consumer the information about the carbon footprint for
    the target service, considering all the active instances, in the period of time of interest


## How to run service locally

TBD

