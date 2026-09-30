-- AuthorizationPost previously created a consent with an empty primary key.
-- An empty key can only exist once, so a reserved legacy ID is sufficient.
SET @legacy_consent_id = '00000000000000000000000000';
SET @previous_foreign_key_checks = @@FOREIGN_KEY_CHECKS;

START TRANSACTION;
SET FOREIGN_KEY_CHECKS = 0;
UPDATE consents
SET id = @legacy_consent_id
WHERE id = '';

UPDATE oauth_tokens
SET consent_id = @legacy_consent_id
WHERE consent_id = '';
COMMIT;
SET FOREIGN_KEY_CHECKS = @previous_foreign_key_checks;
