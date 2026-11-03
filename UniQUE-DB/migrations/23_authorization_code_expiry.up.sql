ALTER TABLE `authorization_requests`
  ADD `code_issued_at` DATETIME(6) NULL DEFAULT NULL AFTER `code`,
  ADD `code_expires_at` DATETIME(6) NULL DEFAULT NULL AFTER `code_issued_at`,
  ADD INDEX `idx_authorization_requests_code_expires_at` (`code_expires_at`);
