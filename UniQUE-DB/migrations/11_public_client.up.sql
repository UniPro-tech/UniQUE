-- Add public_client flag to applications and allow client_secret to be NULL for compatibility
-- Up migration
ALTER TABLE applications MODIFY COLUMN client_secret VARCHAR(255) NULL;

ALTER TABLE applications
ADD COLUMN public_client BOOLEAN NOT NULL DEFAULT FALSE;

-- Note: existing rows will have public_client = FALSE by default.