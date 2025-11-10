**Provider Implementation for the Energy Footprint Notification API**

In this repository a possible implementation of the Transformation Function for the Energy Footprint Notification API is provided.

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
