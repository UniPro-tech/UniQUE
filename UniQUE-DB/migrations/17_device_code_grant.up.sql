ALTER TABLE `authorization_requests` ADD `user_code` CHAR(8) NULL DEFAULT NULL COMMENT 'device flow' AFTER `code`,
ADD `device_code` CHAR(26) NULL DEFAULT NULL COMMENT 'device flow' AFTER `user_code`;

ALTER TABLE `authorization_requests` CHANGE `code` `code` CHAR(26) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL;

ALTER TABLE `authorization_requests` ADD UNIQUE (`device_code`);

ALTER TABLE `authorization_requests` ADD UNIQUE (`user_code`);

ALTER TABLE `authorization_requests` ADD `device_flow_denied` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'device flow' AFTER `is_consented`;

ALTER TABLE `authorization_requests` ADD INDEX `idx_device_code` (`device_code`);

ALTER TABLE `authorization_requests` CHANGE `redirect_uri` `redirect_uri` VARCHAR(255) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  CHANGE `response_type` `response_type` ENUM ('code') CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL;