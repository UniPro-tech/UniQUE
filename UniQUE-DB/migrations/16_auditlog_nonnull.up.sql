ALTER TABLE `sessions` CHANGE `ip_address` `ip_address` VARCHAR(45) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_general_ci NOT NULL;

ALTER TABLE `sessions` CHANGE `user_agent` `user_agent` VARCHAR(225) CHARACTER
SET
  utf8mb4 COLLATE utf8mb4_general_ci NOT NULL;