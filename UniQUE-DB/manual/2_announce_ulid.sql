SET
  GLOBAL log_bin_trust_function_creators = 1;

-- ULID generation function
DELIMITER //

CREATE TRIGGER before_insert_announcements BEFORE INSERT ON announcements FOR EACH ROW BEGIN IF NEW.id IS NULL
OR NEW.id = '' THEN
SET
  NEW.id = gen_ulid ();

END IF;

END;

//

DELIMITER;