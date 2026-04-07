-- users

CREATE TABLE users (
  id CHAR(26) PRIMARY KEY,
  custom_id VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(100) NOT NULL UNIQUE,
  external_email VARCHAR(100) NOT NULL,
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  affiliation_period VARCHAR(20) NULL,
  password_hash VARCHAR(255) NOT NULL,
  status ENUM('established', 'active', 'suspended', 'archived') NOT NULL DEFAULT 'established',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL
) COMMENT='Table to store user credential information. Type: master';

CREATE INDEX idx_users_affiliation_period ON users(affiliation_period);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_updated_at ON users(updated_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- profiles

CREATE TABLE profiles (
  user_id CHAR(26) PRIMARY KEY,
  display_name VARCHAR(100) NOT NULL,
  bio TEXT NULL,
  birthdate DATE NULL,
  twitter_handle VARCHAR(50) NULL,
  website_url VARCHAR(255) NULL,
  joined_at DATE NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='Table to store user profile information. Type: master';

CREATE INDEX idx_profiles_display_name ON profiles(display_name);
CREATE INDEX idx_profiles_joined_at ON profiles(joined_at);
CREATE INDEX idx_profiles_created_at ON profiles(created_at);
CREATE INDEX idx_profiles_updated_at ON profiles(updated_at);
CREATE INDEX idx_profiles_deleted_at ON profiles(deleted_at);

-- applications

CREATE TABLE applications (
  id CHAR(26) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT NULL,
  website_url VARCHAR(255) NULL,
  privacy_policy_url VARCHAR(255) NULL,
  client_secret VARCHAR(255) NOT NULL,
  user_id CHAR(26) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='Table to store third-party application information. Type: master';

CREATE INDEX idx_applications_name ON applications(name);
CREATE INDEX idx_applications_created_at ON applications(created_at);
CREATE INDEX idx_applications_updated_at ON applications(updated_at);
CREATE INDEX idx_applications_deleted_at ON applications(deleted_at);

-- redirect_uris

CREATE TABLE redirect_uris (
  application_id CHAR(26) NOT NULL,
  uri VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
  PRIMARY KEY (application_id, uri)
) COMMENT='Table to store redirect URIs for third-party applications. Type: master';

CREATE INDEX idx_redirect_uris_created_at ON redirect_uris(created_at);
CREATE INDEX idx_redirect_uris_updated_at ON redirect_uris(updated_at);
CREATE INDEX idx_redirect_uris_deleted_at ON redirect_uris(deleted_at);

-- roles

CREATE TABLE roles (
  id CHAR(26) PRIMARY KEY,
  custom_id VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(50) NOT NULL UNIQUE,
  description TEXT NULL,
  permission_bitmask BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL
) COMMENT='Table to store user roles. Type: master';

CREATE INDEX idx_roles_created_at ON roles(created_at);
CREATE INDEX idx_roles_updated_at ON roles(updated_at);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- user_roles
CREATE TABLE user_roles (
  user_id CHAR(26) NOT NULL,
  role_id CHAR(26) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id)
) COMMENT='Table to associate users with roles. Type: transactional';

CREATE INDEX idx_user_roles_created_at ON user_roles(created_at);
CREATE INDEX idx_user_roles_updated_at ON user_roles(updated_at);

-- email_verification_codes

CREATE TABLE email_verification_codes (
  id CHAR(26) PRIMARY KEY,
  user_id CHAR(26) NOT NULL,
  code VARCHAR(255) NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  request_type ENUM('registration', 'password_reset', 'email_change') NOT NULL,
  new_email VARCHAR(100) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='Table to store email verification codes. Type: transactional';

CREATE INDEX idx_email_verification_codes_expires_at ON email_verification_codes(expires_at);
CREATE INDEX idx_email_verification_codes_created_at ON email_verification_codes(created_at);

-- sessions
CREATE TABLE sessions (
  id CHAR(26) PRIMARY KEY,
  user_id CHAR(26) NOT NULL,
  ip_address VARCHAR(45) NULL,
  user_agent VARCHAR(255) NULL,
  expires_at DATETIME NOT NULL,
  last_login_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='Table to store user sessions. Type: master';

CREATE INDEX ids_sessions_ip_address ON sessions(ip_address);
CREATE INDEX idx_sessions_user_agent ON sessions(user_agent);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_login_at ON sessions(last_login_at);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_updated_at ON sessions(updated_at);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- external_identities

CREATE TABLE external_identities (
  id CHAR(26) PRIMARY KEY,
  user_id CHAR(26) NOT NULL,
  provider ENUM('discord', 'github') NOT NULL,
  external_user_id VARCHAR(100) NOT NULL,
  id_token VARCHAR(255) NOT NULL,
  access_token VARCHAR(255) NOT NULL,
  refresh_token VARCHAR(255) NULL,
  token_expires_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='Table to store external identity links for users. Type: master';

CREATE INDEX idx_external_identities_provider ON external_identities(provider);
CREATE INDEX idx_external_identities_external_user_id ON external_identities(external_user_id);
CREATE INDEX idx_external_identities_token_expires_at ON external_identities(token_expires_at);
CREATE INDEX idx_external_identities_created_at ON external_identities(created_at);
CREATE INDEX idx_external_identities_updated_at ON external_identities(updated_at);
CREATE INDEX idx_external_identities_deleted_at ON external_identities(deleted_at);

-- consents

CREATE TABLE consents (
  id CHAR(26) PRIMARY KEY,
  user_id CHAR(26) NOT NULL,
  application_id CHAR(26) NOT NULL,
  scope VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
) COMMENT='Table to store user consents for third-party applications. Type: master';

CREATE INDEX idx_consents_created_at ON consents(created_at);
CREATE INDEX idx_consents_updated_at ON consents(updated_at);
CREATE INDEX idx_consents_deleted_at ON consents(deleted_at);

-- oauth_tokens

CREATE TABLE oauth_tokens (
  id CHAR(26) PRIMARY KEY,
  consent_id CHAR(26) NOT NULL,
  id_token_jti VARCHAR(255) NULL UNIQUE,
  access_token_jti VARCHAR(255) NULL UNIQUE,
  refresh_token_jti VARCHAR(255) NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  FOREIGN KEY (consent_id) REFERENCES consents(id) ON DELETE CASCADE
) COMMENT='Table to store OAuth tokens for third-party applications. Type: master';

CREATE INDEX idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);
CREATE INDEX idx_oauth_tokens_created_at ON oauth_tokens(created_at);
CREATE INDEX idx_oauth_tokens_updated_at ON oauth_tokens(updated_at);
CREATE INDEX idx_oauth_tokens_deleted_at ON oauth_tokens(deleted_at);

-- authorization_requests

CREATE TABLE authorization_requests (
  id CHAR(26) PRIMARY KEY,
  session_id CHAR(26) NULL,
  application_id CHAR(26) NOT NULL,
  scope VARCHAR(255) NOT NULL,
  redirect_uri VARCHAR(255) NOT NULL,
  state VARCHAR(255) NULL,
  nonce VARCHAR(255) NULL,
  prompt ENUM('none', 'login', 'consent', 'select_account') NULL,
  code_challenge VARCHAR(255) NULL,
  code_challenge_method ENUM('plain', 'S256') NULL,
  -- Only 'code' is supported for now
  response_type ENUM('code') NOT NULL DEFAULT 'code',
  is_consented BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Consentが作られており、ユーザーが許可したかどうか',
  code VARCHAR(255) NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
  FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
) COMMENT='Table to store OAuth2 authorization requests. Type: transactional';

CREATE INDEX idx_authorization_requests_expires_at ON authorization_requests(expires_at);
CREATE INDEX idx_authorization_requests_created_at ON authorization_requests(created_at);

-- audit_logs

CREATE TABLE audit_logs (
  id CHAR(26) PRIMARY KEY,
  user_id CHAR(26) NULL,
  application_id CHAR(26) NULL,
  session_id CHAR(26) NULL COMMENT 'If author is application, this links to the jti',
  action ENUM('CREATE', 'READ', 'UPDATE', 'DELETE', 'LOGIN', 'LOGOUT', 'AUTHORIZATION') NOT NULL,
  target_resource VARCHAR(255) NOT NULL,
  trusted BOOLEAN NOT NULL DEFAULT FALSE,
  details TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE SET NULL
) COMMENT='Table to store audit logs for security and compliance. Type: transactional';

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_application_id ON audit_logs(application_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_target_resource ON audit_logs(target_resource);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
