ALTER TABLE `sessions` CHANGE `ip_address` `ip_address` VARCHAR(45) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_general_ci NULL;

ALTER TABLE `sessions` CHANGE `user_agent` `user_agent` VARCHAR(225) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_general_ci NULL;

UPDATE `sessions`
SET
  `ip_address` = NULL
WHERE
  `user_agent` IS 'unknown';

UPDATE `sessions`
SET
  `user_agent` = NULL
WHERE
  `user_agent` IS 'unknown';