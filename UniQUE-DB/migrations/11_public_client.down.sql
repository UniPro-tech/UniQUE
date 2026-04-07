-- Rollback for 11_public_client.up.sql
-- Down migration
ALTER TABLE applications MODIFY COLUMN client_secret VARCHAR(255) NOT NULL;

ALTER TABLE applications
DROP COLUMN public_client;