# Kubernetes deployment

The official chart is located at `deploy/helm/unique`. Tagged releases are
published to Harbor as `oci://registry.uniproject.jp/infra/unique`.

Install a published release with:

```sh
helm upgrade --install unique oci://registry.uniproject.jp/infra/unique \
  --version 0.1.0 \
  --namespace unique \
  --create-namespace \
  --values deploy/helm/unique/values-production.yaml
```

Log in first when the Harbor project is private:

```sh
helm registry login registry.uniproject.jp
```

Every successful `main` build publishes immutable `sha-<commit>` container
tags and a uniquely versioned `0.0.0-main.<run>.<attempt>` development chart.
Release tags publish semantic-version and `latest` container tags. The chart's
`appVersion` selects the matching tag for every UniQUE image by default.

The chart expects the existing application Secrets and `internal-harbor` image
pull Secret. Apply the encrypted manifests from
`deploy/sealed-secrets/production` before installing or upgrading the chart:

```sh
kubectl apply -f deploy/sealed-secrets/production
```

Validate local changes with:

```sh
helm lint deploy/helm/unique
helm template unique deploy/helm/unique --namespace unique
helm lint deploy/helm/unique --values deploy/helm/unique/values-production.yaml
```

The production values enable the existing Gateway API routes. Disable
`networkPolicy` or `ciliumNetworkPolicy` only when the cluster does not support
the corresponding policy implementation.
