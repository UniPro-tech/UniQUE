ALTER TABLE users
ADD COLUMN `avatar` VARCHAR(20) NOT NULL DEFAULT 'email' AFTER `email_verified`,
ADD CONSTRAINT chk_avatar
CHECK (avatar IN ('external_email', 'email', 'upload'));