# 本番環境のSealedSecret

暗号化済みリソースは環境固有のため、意図的にHelm Chartの外で管理しています。
UniQUEをインストールする前に、次のコマンドで適用してください。

```sh
kubectl apply -f deploy/sealed-secrets/production
```

`cert.pem` は暗号化に使用する公開証明書です。秘密鍵は絶対にコミットしないでください。

## Database DSN keys

Before upgrading the chart, add `MYSQL_NATIVE_DSN` to the Secret referenced by
`database.existingSecret` (`mysql-secret` in production). Mail and Discord use it
for settings queries with the Go MySQL driver:

```text
user:pass@tcp(host:3306)/dbname
```

Use the existing database credentials and database name, with the in-cluster
MySQL service as the host. Seal this new value with `cert.pem` and add it to
`production/mysql-secret.yaml` under `spec.encryptedData`; do not commit the
plaintext DSN. The checked-in SealedSecret still needs this new encrypted key.
Keep `MYSQL_ACCESS_URI` unchanged: migrations continue to use its `mysql://` URI
format.
