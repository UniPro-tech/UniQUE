# Kubernetesデプロイ

公式Chartは `deploy/helm/unique` にあります。タグ付きリリースは、
`oci://registry.uniproject.jp/infra/unique` としてHarborへ公開されます。

`deploy/helm/unique/values-production.yaml` はリポジトリ内のローカルファイルです。
コマンドを実行する前にリポジトリを clone し、ルートディレクトリへ移動してください。

```sh
git clone https://github.com/UniPro-tech/UniQUE.git
cd UniQUE
```

公開済みのリリースは、次のコマンドでインストールできます。

```sh
helm upgrade --install unique oci://ghcr.io/unipro-tech/charts/unique \
  --version 0.1.0 \
  --namespace unique \
  --create-namespace \
  --wait --wait-for-jobs --timeout 15min \
  --values deploy/helm/unique/values-production.yaml
```

Harborのプロジェクトが非公開の場合は、先にログインしてください。

```sh
helm registry login registry.uniproject.jp
```

`main` のビルドが成功するたびに、変更されない `sha-<commit>` コンテナタグと、
一意なバージョン `0.0.0-main.<run>.<attempt>` を持つ開発用Chartが公開されます。
リリースタグでは、セマンティックバージョンと `latest` のコンテナタグが公開されます。
デフォルトでは、Chartの `appVersion` と一致するタグがすべてのUniQUEイメージに使用されます。

ローカルでの変更は、次のコマンドで検証できます。

```sh
helm lint deploy/helm/unique
helm template unique deploy/helm/unique --namespace unique
helm lint deploy/helm/unique --values deploy/helm/unique/values-production.yaml
```

本番用Valuesでは、既存のGateway APIルートが有効になります。クラスタが対応する
ポリシー実装をサポートしていない場合に限り、`networkPolicy` または
`ciliumNetworkPolicy` を無効にしてください。

インストール・更新時には通常の Job でマイグレーションを実行し、MySQL が利用可能になるまで
再試行します。API・Auth・Mail・Discord は init container で、イメージに同梱された
最新のスキーマバージョンへの移行が完了し、dirty 状態でないことを確認してから起動します。
マイグレーションが失敗した場合は起動を保留します。日次の CronJob は使用しません。
