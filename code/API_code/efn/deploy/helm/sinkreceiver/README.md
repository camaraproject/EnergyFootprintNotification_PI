# sinkreceiver Helm Chart

This chart deploys the sinkreceiver service with HTTPS support for internal cluster communication.

## How it works
- Deploys a ClusterIP service exposing port 8090 (HTTP, internal) and 8443 (HTTPS).
- The sinkreceiver Go application listens on port 8090 (HTTP only).
- An nginx sidecar container (`tls-proxy`) terminates TLS on port 8443 and proxies to port 8090.
- A self-signed TLS certificate is automatically provisioned by cert-manager with proper Subject Alternative Names (SANs).
- The Notification service can use `https://sinkreceiver.<namespace>.svc.cluster.local:8443` as the callback sink URL.

## Prerequisites
- [cert-manager](https://cert-manager.io/) installed in your cluster.

## Usage
1. Install cert-manager if not already present:
   ```sh
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```
2. Deploy this chart:
   ```sh
   helm install sinkreceiver ./deploy/helm/sinkreceiver --namespace <your-namespace>
   ```
3. Wait for the certificate to be ready:
   ```sh
   kubectl wait --for=condition=Ready certificate/sinkreceiver-tls -n <your-namespace> --timeout=60s
   ```
4. Use the following sink URL in your subscription request:
   ```
   https://sinkreceiver.<your-namespace>.svc.cluster.local:8443
   ```

## Certificate Details
- The chart creates a self-signed ClusterIssuer for internal cluster use.
- The certificate includes multiple SANs to support various DNS resolution patterns.
- The certificate is valid for 1 year and auto-renewed 30 days before expiration.
- Modern Go TLS clients (Go 1.15+) require SANs; the CN-only approach is deprecated.

## Trusting the Certificate
For the notification service to successfully connect via HTTPS, it needs to trust the self-signed CA. You have two options:

### Option 1: Use the Certificate in the notification service
Mount the same certificate secret and configure the notification service to trust it.

### Option 2: Skip verification (NOT recommended for production)
Configure the notification HTTP client to skip certificate verification (for testing only).

## Notes
- The certificate is self-signed and suitable for internal cluster communication.
- For external access or production environments with strict security requirements, consider using a proper CA.
- The nginx sidecar handles all TLS termination; the Go application remains unchanged.
