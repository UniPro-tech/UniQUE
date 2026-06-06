# UniQUE - Open ID Connect ID Provider

このプロジェクトは、[デジタル創作サークルUniProject](https://uniproject.jp)のためのOIDC認証基盤です。
Discord連携機能やアナウンス機能などを持ち合わせており、サークルの円滑な運営のために用いられています。

## Components

このプロジェクトはマイクロサービスアーキテクチャの思想を取り入れ、主に5つのコンポーネントに分離しています。

### Resource API Server

Goで作られたリソースAPIサーバーです。
ここで、UniQUEに保存されたすべてのデータを管理できます。

### Auth Server

認証・認可サーバーです。
ここでは、Open ID Connect 1.0 / OAuth 2.0に基づいた処理を行っています。
また、フロントエンドから受け取る認証リクエストを処理しています。

### Frontend

フロントエンドです。
ユーザーは基本的のこのコンポーネントを経由しWebUIを用いてアクセスします。
Next.jsで制作されており、このコンポーネントからデータベースへのアクセスは許可されていません。

### Mail Server

メールを送るためだけのシステムです。
ここであらゆる種類のメールのテンプレートと変換し、APIサーバー等から送信されたリクエストをもとにメールを配信します。

### UniQUE-DB

これは、マイグレーションファイルが含まれたデータベースマイグレーション用のコンポーネントです。
KubernetesのCronJobにより、1日1回、GitからCloneしたマイグレーションファイルをデータベースに反映します。

また、DBはMySQL互換です。

## Deployment

デプロイにはKustomizationを用いています。
kuatomizationsディレクトリ内のファイルをArgoCDを用いてデプロイしています。

### Sealed Secretについて

暗号化のためには、下記条件を指定してください。

- mode - strict
- namespace - unique
- public key - cert.pem

## Contributing & Development

コントリビュートに興味をお持ちいただき、ありがとうございます！

開発環境はdocker-composeを用いて整えることができます。

### Tag

タグについては、下記のルールに従いましょう。

- [semver 2.0](https://semver.org/lang/ja/)を採用しています。
- 下記のprefixを必ずつけましょう
  - front/v
  - auth/v
  - api/v
  - mail/v
  - db/v

## LICENSE

このプロジェクトは、AGPL-3.0ライセンスの下で公開されています。詳しくは、LICENSEファイルをご覧ください。

## Super Thanks

このプロジェクトにおいて、大きな変化をもたらしてくださった方々です。
本当にありがとうございました。

- @sibapybot - サーバーを提供していただき、多くのアイデアを提供してくださいました。

その他すべてのコントリビューターおよびUniProjectメンバーに感謝します。

### 以前のリポジトリのコントリビューター

<a href="https://github.com/UniPro-tech/UniQUE-Auth/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=UniPro-tech/UniQUE-Auth" />
</a>

<a href="https://github.com/UniPro-tech/UniQUE-API/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=UniPro-tech/UniQUE-API" />
</a>

<a href="https://github.com/UniPro-tech/UniQUE-Front/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=UniPro-tech/UniQUE-Front" />
</a>

## Old Repositories

このリポジトリはmonorepoになる前の分離されたバージョンが存在します。
過去のバージョンを追うには下記リポジトリを参照してください。

- [UniQUE-Auth](https://github.com/UniPro-tech/UniQUE-Auth)
- [UniQUE-API](https://github.com/UniPro-tech/UniQUE-API)
- [UniQUE-Front](https://github.com/UniPro-tech/UniQUE-Front)
- [UniQUE-MailServer](https://github.com/UniPro-tech/UniQUE-MailServer)
- [UniQUE-DB](https://github.com/UniPro-tech/UniQUE-DB)
- [UniQUE-Kustomization](https://github.com/UniPro-tech/UniQUE-Kustomization)
