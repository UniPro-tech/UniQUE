CREATE TABLE
  `announcements` (
    `id` CHAR(26) PRIMARY KEY,
    `title` VARCHAR(255) NOT NULL,
    `content` text NOT NULL,
    `created_by` CHAR(26) NOT NULL,
    `is_pinned` BOOLEAN NOT NULL DEFAULT FALSE,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE NO ACTION ON UPDATE CASCADE
  );

CREATE INDEX `idx_announcements_created_by` ON `announcements` (`created_by`);

CREATE INDEX `idx_announcements_is_pinned` ON `announcements` (`is_pinned`);

CREATE INDEX `idx_announcements_deleted_at` ON `announcements` (`deleted_at`);

CREATE INDEX `idx_announcements_created_at` ON `announcements` (`created_at`);

CREATE INDEX `idx_announcements_updated_at` ON `announcements` (`updated_at`);

CREATE INDEX `idx_announcements_title` ON `announcements` (`title`);