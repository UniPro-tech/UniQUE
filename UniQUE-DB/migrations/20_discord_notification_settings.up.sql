CREATE TABLE discord_notification_settings (
  id TINYINT UNSIGNED PRIMARY KEY,
  channel_id VARCHAR(20) NOT NULL,
  notify_announcements BOOLEAN NOT NULL DEFAULT TRUE,
  notify_registration_requests BOOLEAN NOT NULL DEFAULT TRUE,
  notify_migrations BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT chk_discord_notification_settings_singleton CHECK (id = 1)
) COMMENT='Singleton settings for Discord notifications. Type: master';
