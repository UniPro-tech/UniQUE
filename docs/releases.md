# リリース

UniQUEでは、Conventional Commitsからリリースを準備するためにRelease Pleaseを使用します。
`main` へ変更がpushされると、次の内容を含む1件のリリースPull Requestが作成または更新されます。

- `.release-please-manifest.json` の次期セマンティックバージョン
- `deploy/helm/unique/Chart.yaml` の対応する `version` と `appVersion`
- `CHANGELOG.md` に生成されるリリースノート

リリースPull Requestをマージすると、対応するGitHub Releaseと `vX.Y.Z` タグが作成されます。
タグを契機にワークフローが実行され、バージョン付きの全6サービスのコンテナイメージと
OCI Helm Chartが公開されます。**リリースタグを手動で作成しないでください**。

## バージョンの決定方法

- `fix:` はパッチリリースになります。
- `feat:` はマイナーリリースになります。
- `feat!:`、`fix!:`、または `BREAKING CHANGE:` フッターを含む変更はメジャーリリースになります。

その他のConventional Commitタイプは、バージョンを変更せずリリースノートに含まれる場合があります。
リリース可能なコミットが `main` に追加されるたびに、開いているリリースPull Requestが更新されます。
そのPull Requestのマージを、公開の明示的な承認として扱います。
