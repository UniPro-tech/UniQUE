# リリース

UniQUEでは、Conventional Commitsからリリースを準備するためにRelease Pleaseを使用します。
`main` へ変更がpushされると、次の内容を含む1件のリリースPull Requestが作成または更新されます。

- `.release-please-manifest.json` の次期セマンティックバージョン
- `deploy/helm/unique/Chart.yaml` の対応する `version` と `appVersion`
- `CHANGELOG.md` に生成されるリリースノート

リリースPull Requestをマージすると、対応するGitHub Releaseと `vX.Y.Z` タグが作成されます。
タグを契機にワークフローが実行され、バージョン付きの全6サービスのコンテナイメージと
OCI Helm Chartが公開されます。リリースタグを手動で作成しないでください。

## リポジトリの設定

リリース自動化用のGitHub Appを、次のリポジトリ権限で作成します。

- Contents: Read and write
- Issues: Read and write
- Pull requests: Read and write
- Metadata: Read-only

Appをこのリポジトリへインストールして秘密鍵を生成し、次の値を設定します。

- Actions変数 `RELEASE_APP_CLIENT_ID`: AppのクライアントID
- Actions Secret `RELEASE_APP_PRIVATE_KEY`: PEM形式の秘密鍵全文
- Actions Secret `HARBOR_TOKEN`: `robot$github-publisher` のパスワード

組み込みの `GITHUB_TOKEN` で作成したPull Requestやタグでは、後続のGitHub Actions
ワークフローが起動しないため、GitHub Appのトークンが必要です。

## バージョンの決定方法

- `fix:` はパッチリリースになります。
- `feat:` はマイナーリリースになります。
- `feat!:`、`fix!:`、または `BREAKING CHANGE:` フッターを含む変更はメジャーリリースになります。

その他のConventional Commitタイプは、バージョンを変更せずリリースノートに含まれる場合があります。
リリース可能なコミットが `main` に追加されるたびに、開いているリリースPull Requestが更新されます。
そのPull Requestのマージを、公開の明示的な承認として扱います。
