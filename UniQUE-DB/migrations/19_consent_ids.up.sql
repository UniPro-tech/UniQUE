-- AuthorizationPost previously created a consent with an empty primary key.
-- An empty key can only exist once, so a reserved legacy ID is sufficient.
SET @legacy_consent_id = '00000000000000000000000000';

SET FOREIGN_KEY_CHECKS = 0;
UPDATE consents
SET id = @legacy_consent_id
WHERE id = '';

UPDATE oauth_tokens
SET consent_id = @legacy_consent_id
WHERE consent_id = '';
SET FOREIGN_KEY_CHECKS = 1;
