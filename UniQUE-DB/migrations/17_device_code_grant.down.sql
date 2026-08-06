ALTER TABLE `authorization_requests`
DROP `user_code`,
DROP `device_code`,
DROP `device_flow_denied`;

ALTER TABLE `authorization_requests` CHANGE `code` `code` VARCHAR(255) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL;

ALTER TABLE `authorization_requests` CHANGE `redirect_uri` `redirect_uri` VARCHAR(255) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  CHANGE `response_type` `response_type` ENUM ('code') CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'code';