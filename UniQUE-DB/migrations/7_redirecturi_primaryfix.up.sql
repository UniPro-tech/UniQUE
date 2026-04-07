ALTER TABLE `redirect_uris`
DROP FOREIGN KEY `redirect_uris_ibfk_1`;
ALTER TABLE `redirect_uris`
DROP PRIMARY KEY;
ALTER TABLE `redirect_uris` ADD CONSTRAINT `redirect_uris_ibfk_1` FOREIGN KEY (`application_id`) REFERENCES `applications`(`id`) ON DELETE CASCADE ON UPDATE CASCADE;
CREATE INDEX `idx_redirect_uris_uri` ON `redirect_uris` (`uri`);