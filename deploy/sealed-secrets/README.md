# 本番環境のSealedSecret

暗号化済みリソースは環境固有のため、意図的にHelm Chartの外で管理しています。
UniQUEをインストールする前に、次のコマンドで適用してください。

```sh
kubectl apply -f deploy/sealed-secrets/production
```

`cert.pem` は暗号化に使用する公開証明書です。秘密鍵は絶対にコミットしないでください。
