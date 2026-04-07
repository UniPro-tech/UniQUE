ALTER TABLE `users`
DROP COLUMN `totp_secret`;

ALTER TABLE `users`
DROP COLUMN `is_totp_enabled`;