# Application settings

UniQUE stores administrator-managed, non-connection configuration in the
`settings` table. Configuration is resolved in this order:

1. an explicitly set environment variable;
2. the corresponding database setting;
3. the built-in default, when one exists.

Database and service connection values such as `DB_DSN`, frontend/API/issuer
URLs, and key file paths remain environment variables so the services can
connect before reading settings.

Users with the `CONFIG_UPDATE` permission can manage the allowlisted keys with:

- `GET /settings`
- `GET /settings/{key}`
- `PUT /settings/{key}` with `{ "value": "..." }`
- `DELETE /settings/{key}`

Secret values such as Discord tokens, OAuth client secrets, and SMTP passwords
are write-only: list and read responses report whether they are configured but
never return the value. Responses also report when an environment variable is
overriding the stored value.

The existing `/settings/discord-notifications` API now writes the notification
channel and event switches into the same table. Migration 21 copies existing
Issue #28 values without deleting the old compatibility table.

API settings are loaded at process startup; Discord notification settings are
read for each notification. Restart Auth, Mail, Discord, or API pods after
changing another setting. In Kubernetes, Mail and Discord receive
`SETTINGS_DB_DSN` from the existing `mysql-secret`; the database URI itself is
not stored in the settings table.
