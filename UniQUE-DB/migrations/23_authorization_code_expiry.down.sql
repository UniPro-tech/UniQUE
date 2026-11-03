ALTER TABLE `authorization_requests`
  DROP INDEX `idx_authorization_requests_code_expires_at`,
  DROP `code_expires_at`,
  DROP `code_issued_at`;
