INSERT INTO
  `users` (
    `id`,
    `custom_id`,
    `email`,
    `external_email`,
    `email_verified`,
    `affiliation_period`,
    `password_hash`,
    `totp_secret`,
    `is_totp_enabled`,
    `status`
  )
VALUES
  (
    '01KWMB8BJ6W895W6SJ6NMJSV17',
    'test',
    'test@uniproject.jp',
    'test@test.com',
    '1',
    '00',
    '$2a$12$DoQvVJqm40s8iURbyPD11eiUiasjSo/LvnJhakdGb3dNNW4wpLoSi',
    NULL,
    '0',
    'active'
  );

INSERT INTO
  `profiles` (`user_id`, `display_name`, `bio`, `birthdate`, `birthdate_visible`, `twitter_handle`, `website_url`, `joined_at`)
VALUES
  ('01KWMB8BJ6W895W6SJ6NMJSV17', 'test', NULL, NULL, '0', NULL, NULL, NULL);

INSERT INTO
  `users` (
    `id`,
    `custom_id`,
    `email`,
    `external_email`,
    `email_verified`,
    `affiliation_period`,
    `password_hash`,
    `totp_secret`,
    `is_totp_enabled`,
    `status`
  )
VALUES
  (
    '01KWMB8BJ6W895W6SJ6NMJSV18',
    'test-user',
    'test-user@uniproject.jp',
    'test-user@test.com',
    '1',
    '00',
    '$2a$12$DoQvVJqm40s8iURbyPD11eiUiasjSo/LvnJhakdGb3dNNW4wpLoSi',
    NULL,
    '0',
    'active'
  );

INSERT INTO
  `profiles` (`user_id`, `display_name`, `bio`, `birthdate`, `birthdate_visible`, `twitter_handle`, `website_url`, `joined_at`)
VALUES
  ('01KWMB8BJ6W895W6SJ6NMJSV18', 'test-user', NULL, NULL, '0', NULL, NULL, NULL);

INSERT INTO
  `roles` (`id`, `custom_id`, `name`, `description`, `permission_bitmask`, `is_default`)
VALUES
  ('01KH99W452006KZ5EB0VCZN38N', 'admin', '役員', 'UniProjectの役員です。', '-1', '0'),
  ('01KH99W452006KZ5EB0VCZN38P', 'member', '一般ユーザー', '一般ユーザーの権限です。', '0', '1');

INSERT INTO
  `user_roles` (`user_id`, `role_id`)
VALUES
  ('01KWMB8BJ6W895W6SJ6NMJSV17', '01KH99W452006KZ5EB0VCZN38N'),
  ('01KWMB8BJ6W895W6SJ6NMJSV18', '01KH99W452006KZ5EB0VCZN38P')