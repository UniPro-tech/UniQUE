SET GLOBAL log_bin_trust_function_creators = 1;

-- ULID generation function

DELIMITER //

DELIMITER //

CREATE FUNCTION to_crockford_b32 (src BIGINT, encoded_len INT)
RETURNS CHAR(26)
DETERMINISTIC
NO SQL
BEGIN
    DECLARE result TEXT DEFAULT '';
    DECLARE b32char CHAR(32) DEFAULT '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    DECLARE i INT DEFAULT 0;

    WHILE i < encoded_len DO
        SET result = CONCAT(
            SUBSTRING(b32char, (src MOD 32) + 1, 1),
            result
        );
        SET src = src DIV 32;
        SET i = i + 1;
    END WHILE;

    RETURN result;
END//

CREATE FUNCTION gen_ulid ()
RETURNS CHAR(26)
NOT DETERMINISTIC
NO SQL
BEGIN
    DECLARE msec_ts BIGINT;
    DECLARE rand_hex CHAR(32);
    DECLARE rand_hi BIGINT;
    DECLARE rand_lo BIGINT;

    SET msec_ts = FLOOR(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000);
    SET rand_hex = HEX(RANDOM_BYTES(10));

    SET rand_hi = CONV(SUBSTRING(rand_hex, 1, 8), 16, 10);
    SET rand_lo = CONV(SUBSTRING(rand_hex, 9, 12), 16, 10);

    RETURN CONCAT(
        to_crockford_b32(msec_ts, 10),
        to_crockford_b32(rand_hi, 8),
        to_crockford_b32(rand_lo, 8)
    );
END//
DELIMITER ;

DELIMITER //
CREATE TRIGGER before_insert_users
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
      SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_applications
BEFORE INSERT ON applications
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_roles
BEFORE INSERT ON roles
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_email_verification_codes
BEFORE INSERT ON email_verification_codes
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_sessions
BEFORE INSERT ON sessions
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_external_identities
BEFORE INSERT ON external_identities
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_consents
BEFORE INSERT ON consents
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_oauth_tokens
BEFORE INSERT ON oauth_tokens
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_audit_logs
BEFORE INSERT ON audit_logs
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
CREATE TRIGGER before_insert_authorization_requests
BEFORE INSERT ON authorization_requests
FOR EACH ROW
BEGIN
  IF NEW.id IS NULL OR NEW.id = '' THEN
    SET NEW.id = gen_ulid();
  END IF;
END; //
DELIMITER ;