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

開発用バックエンドはDocker Composeで起動します。本番のKustomize/SealedSecretは対象外です。

1. 各サンプル環境ファイルをローカル用ファイルとしてコピーします。

   ```sh
   cp UniQUE-DB/.env.example UniQUE-DB/.env
   cp UniQUE-API/.env.example UniQUE-API/.env
   cp UniQUE-Auth/.env.example UniQUE-Auth/.env
   cp UniQUE-MailServer/.env.example UniQUE-MailServer/.env
   cp UniQUE-Front/.env.example UniQUE-Front/.env
   cp UniQUE-Front/.env.compose.example UniQUE-Front/.env.compose
   ```

2. Discord、SMTP、GAS連携を試す場合は、対応するローカル環境ファイルにテスト用の値を設定します。
3. バックエンドを起動します。

   ```sh
   docker compose up -d --build
   ```

4. フロントエンドは通常、ホスト上でBunを使って起動します。

   ```sh
   cd UniQUE-Front
   bun install
   bun run dev
   ```

5. Compose上でフロントエンドも含めて確認する場合は、次を実行します。ソース変更は自動反映されます。

   ```sh
   docker compose --profile frontend up -d --build
   ```

6. phpMyAdminにアクセスしてUniQUE-DB/manual内のsqlファイルを順番に実行します。
   phpMyAdminのURLは `http://localhost:3001` です。
   また、デフォルトでは、DBの名前はdevdbです。

7. OIDCのAuthorization Code Flowを確認する場合は、テストクライアント用の環境ファイルも作成して起動します。

   ```sh
   cp dev/oidc-client/.env.example dev/oidc-client/.env
   docker compose --profile oidc up -d --build
   ```

   `http://localhost:3002` からログインを開始できます。

   プロトコルの自動確認は次で実行します。

   ```sh
   bun --cwd dev/oidc-client run test:oidc
   ```

### テストユーザー

上記の手順6により開発用ユーザーを冪等に作成します。

- User: `test`
- Password: `testpassword`

### シードデータについて

認証情報は下記の通りです。

- User: `test`
- PW: `testpassword`

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
