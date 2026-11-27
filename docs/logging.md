# ログスキーマ

UniQUE の各サービス（API / Auth / MailServer / Discord / Front）は、
**JSON 形式の構造化ログ**を出力します。バックエンドは Go `log/slog`、
フロントエンドは `src/libs/logger.ts` を使用します。この文書は全サービスで
共通のスキーマを定めます。

## 基本

- バックエンド: `slog.NewJSONHandler(os.Stdout, ...)`
- フロントエンド: `logger.info/warn/error/debug` が JSON 行を出力
- 既定レベル: `Info`
- 既定ロガーに `component` を必ず付与する
  （`api` / `auth` / `mail` / `discord` / `front`）
- `msg` は機械可読な **snake_case のイベント名**（例: `http_request`,
  `discord_api_error`）。文章を入れない。
- 人が読む文字列は `message` に入れる（任意）。
- キーは **snake_case**。値は文字列・数値・真偽値・JSON オブジェクトのいずれか。

## 相関 ID

| キー | 型 | 説明 |
| --- | --- | --- |
| `request_id` | string | 1 リクエスト単位の ID（32 桁 hex） |
| `trace_id` | string | 分散トレース ID（W3C、32 桁 hex） |

- 受信時は `X-Request-ID` / `X-Trace-ID` / `traceparent` を読み、無ければ生成する。
- 応答ヘッダーにも `X-Request-ID` / `X-Trace-ID` を返す。
- サービス間・外部 API 呼び出しでは `X-Request-ID` / `X-Trace-ID` /
  `traceparent` を引き継ぐ。
- フロントの API 呼び出し（`src/libs/apiClient.ts`）も同じヘッダーを送る。

## 共通キー

| キー | 型 | 説明 |
| --- | --- | --- |
| `component` | string | サービス名（必須） |
| `subsystem` | string | サービス内の機能名（任意。例: `discord`） |
| `operation` | string | 論理操作名（任意。例: `add_guild_member`） |
| `message` | string | 人が読むメッセージ（任意） |
| `error` | string / object | エラー。`slog.Any("error", err)` を使う |
| `status` | int | HTTP / 外部 API のステータス |
| `latency_ms` | int | リクエスト処理時間（ミリ秒） |
| `method` | string | HTTP メソッド |
| `path` | string | リクエストパス |
| `query` | string | クエリ文字列（**機密値はマスク済み**） |
| `ip` | string | クライアント IP |
| `user_agent` | string | User-Agent |
| `user_id` | string | UniQUE ユーザー ID |
| `session_id` | string | セッション ID（**マスク済み**） |
| `role_id` | string | ロール ID |
| `application_id` | string | アプリケーション ID |
| `discord_user_id` | string | Discord ユーザー ID |
| `guild_id` | string | Discord ギルド ID |
| `channel_id` | string | Discord チャンネル ID |
| `response_body` | string | 外部 API のレスポンス本文（**マスク・切り詰め済み**） |

## レベル

| レベル | 用途 |
| --- | --- |
| `Debug` | 成功時の詳細（外部 API の正常応答、権限読み込みなど） |
| `Info` | サービス起動・正常な HTTP リクエスト |
| `Warn` | 復旧可能な失敗・外部サービスの異常・認証拒否 |
| `Error` | 処理を継続できない失敗・panic |

## 機密情報

以下の型（`slog.LogValuer`）を使ってマスクする。

- `Secret` — トークン・セッション ID など。常に `[REDACTED]`
- `MaskedBody` — JSON 本文。`access_token` / `refresh_token` / `id_token` /
  `secret` / `client_secret` / `bot_token` / `password` / `authorization` /
  `api_key` / `jti` / `sid` / `session_id` を再帰的にマスク。2048 バイトで切り詰め
- `RedactedURL` — URL / クエリ文字列。`token` / `code` / `state` / `jti` /
  `session_id` などをマスク

## 主なイベント名

- ライフサイクル: `service_starting`, `service_start_failed`,
  `database_init_failed`, `settings_load_failed`, `panic_recovered`
- HTTP: `http_request`, `http_error`
- ライブラリ出力: `gin_output`, `gorm_info`, `gorm_warning`, `gorm_error`,
  `gorm_query`, `gorm_query_failed`
- フロント: `api_request`, `api_response`, `api_error`,
  `front_info` / `front_warn` / `front_error` / `front_debug`
  （`message` に元の文言、`msg` は文言から生成した slug）
- 認証: `jwt_parse_failed`, `jwt_issuer_mismatch`, `token_expired`,
  `token_verification_failed`, `session_verify_failed`
- 認可・管理: `role_assign_failed`, `audit_log_insert_failed`
- Discord: `discord_api_response`, `discord_api_error`,
  `discord_guild_join_failed`, `discord_role_assign_failed`,
  `discord_welcome_dm_failed`, `discord_notification_failed`
