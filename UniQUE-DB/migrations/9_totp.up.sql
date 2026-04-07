ALTER TABLE `users`
ADD COLUMN `totp_secret` VARCHAR(255) DEFAULT NULL AFTER `password_hash`;

ALTER TABLE `users`
ADD COLUMN `is_totp_enabled` BOOLEAN NOT NULL DEFAULT FALSE AFTER `totp_secret`;