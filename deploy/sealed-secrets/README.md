# Production SealedSecrets

These encrypted resources are intentionally kept outside the Helm chart because
they are environment-specific. Apply them before installing UniQUE:

```sh
kubectl apply -f deploy/sealed-secrets/production
```

`cert.pem` is the public sealing certificate. Never commit the private key.
