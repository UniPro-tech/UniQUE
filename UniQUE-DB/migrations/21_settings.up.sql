CREATE TABLE settings (
  `key` VARCHAR(100) PRIMARY KEY,
  `value` TEXT NOT NULL,
  is_secret BOOLEAN NOT NULL DEFAULT FALSE,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT='Application settings. Environment variables take precedence over these values.';

CREATE INDEX idx_settings_updated_at ON settings(updated_at);

-- Preserve the Issue #28 table for backwards compatibility while copying its
-- values into the general-purpose settings store.
INSERT INTO settings (`key`, `value`, is_secret, description)
SELECT 'discord.notification_channel_id', channel_id, FALSE, 'Discord notification channel ID'
FROM discord_notification_settings
WHERE id = 1;

INSERT INTO settings (`key`, `value`, is_secret, description)
SELECT 'discord.notify_announcements', IF(notify_announcements, 'true', 'false'), FALSE, 'Notify when an announcement is created'
FROM discord_notification_settings
WHERE id = 1;

INSERT INTO settings (`key`, `value`, is_secret, description)
SELECT 'discord.notify_registration_requests', IF(notify_registration_requests, 'true', 'false'), FALSE, 'Notify when a registration is requested'
FROM discord_notification_settings
WHERE id = 1;

INSERT INTO settings (`key`, `value`, is_secret, description)
SELECT 'discord.notify_migrations', IF(notify_migrations, 'true', 'false'), FALSE, 'Notify when an account migration completes'
FROM discord_notification_settings
WHERE id = 1;
