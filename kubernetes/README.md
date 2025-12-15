# IronBuckets Kubernetes Deployment

This directory contains Kubernetes manifests for deploying IronBuckets using
[Flux](https://fluxcd.io/) and the
[bjw-s app-template](https://github.com/bjw-s/helm-charts/tree/main/charts/library/common)
Helm chart.

## Architecture

The deployment runs two containers in a single pod:

- **ironbuckets** - The main web UI application (port 8080)
- **minio** - MinIO object storage server (ports 9000, 9001)

## Container Images

Images are published to GitHub Container Registry on each release:

- `ghcr.io/damacus/ironbuckets:latest`
- `ghcr.io/damacus/minio-community:latest`

Both images support `linux/amd64` and `linux/arm64` architectures.

## Prerequisites

- Kubernetes cluster with Flux installed
- [bjw-s HelmRepository](https://github.com/bjw-s/helm-charts) configured
- External Secrets Operator (optional, for secret management)
- Gateway API or Ingress controller

## Directory Structure

```text
kubernetes/
└── apps/
    └── ironbuckets/
        ├── ks.yaml                    # Flux Kustomization
        └── app/
            ├── HelmRelease.yaml       # App-template HelmRelease
            ├── externalsecret.yaml    # External secret for credentials
            └── kustomization.yaml     # Kustomize resources
```

## Configuration

### Required Secrets

Create a secret named `ironbuckets-secret` with:

- `MINIO_ACCESS_KEY` - MinIO access key / username
- `MINIO_SECRET_KEY` - MinIO secret key / password

If using External Secrets Operator, update `externalsecret.yaml` with your
secret store configuration.

### Manual Secret Creation

```bash
kubectl create secret generic ironbuckets-secret \
  --from-literal=MINIO_ACCESS_KEY=your-access-key \
  --from-literal=MINIO_SECRET_KEY=your-secret-key
```

### Customization

Edit `HelmRelease.yaml` to customize:

- **Image tags** - Pin to specific versions instead of `latest`
- **Resource limits** - Adjust CPU/memory based on your needs
- **Persistence** - Configure storage class and size
- **Routes/Ingress** - Update hostnames for your domain

## Standalone Deployment

If not using Flux, you can deploy with Helm directly:

```bash
helm repo add bjw-s https://bjw-s.github.io/helm-charts
helm repo update

# Extract values from HelmRelease.yaml and apply
helm install ironbuckets bjw-s/app-template \
  --values <your-values.yaml>
```

## Services

The deployment creates three services:

| Service                    | Port | Description              |
| -------------------------- | ---- | ------------------------ |
| ironbuckets-app            | 8080 | IronBuckets web UI       |
| ironbuckets-minio-api      | 9000 | MinIO S3 API             |
| ironbuckets-minio-console  | 9001 | MinIO web console        |

## Health Checks

Both containers have liveness and readiness probes configured:

- **IronBuckets**: `GET /health` on port 8080
- **MinIO**: `GET /minio/health/live` and `/minio/health/ready` on port 9000
