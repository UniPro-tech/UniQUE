-- audit_logs table rollback
DROP TABLE IF EXISTS audit_logs;

-- authorizations table rollback
DROP TABLE IF EXISTS authorization_requests;

-- oauth tokens table rollback
DROP TABLE IF EXISTS oauth_tokens;

-- consents table rollback
DROP TABLE IF EXISTS consents;

-- external_identities table rollback
DROP TABLE IF EXISTS external_identities;

-- sessions table rollback
DROP TABLE IF EXISTS sessions;

-- email_verification_codes table rollback
DROP TABLE IF EXISTS email_verification_codes;

-- user_roles
DROP TABLE IF EXISTS user_roles;

-- roles
DROP TABLE IF EXISTS roles;

-- redirect_uris table rollback
DROP TABLE IF EXISTS redirect_uris;

-- applications table rollback
DROP TABLE IF EXISTS applications;

-- profiles
DROP TABLE IF EXISTS profiles;

-- users table rollback
DROP TABLE IF EXISTS users;

-- delete function
DROP FUNCTION IF EXISTS gen_ulid();