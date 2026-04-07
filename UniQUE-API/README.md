# UniQUE API

デジタル創作サークルUniProject 内製認証基盤 UniQUE のリソースAPIサーバーです。

## How to use

> [!NOTE]
> WIP

## 開発方針

1. Many to Manyのリレーションについては下位リソースをネストしない
2. One to Menyのリレーションについては下位リソースをネストする
3. なるべくDB通りの命名を心がける(Authサーバー側との競合を避けるため)

## Swagger

```shell
swag init --generalInfo cmd/server/main.go --parseDependency --parseInternal
```
