-- EasyMail v2 baseline migration placeholder
-- Keep this file idempotent for big-bang cutover.

-- accounts/domains/admin/emails/configures tables are managed by GORM AutoMigrate in current codebase.
-- Add forward-compatible indexes for hot paths.

CREATE INDEX IF NOT EXISTS idx_emails_account_folder_deleted ON emails (account_id, folder_id, deleted);
CREATE INDEX IF NOT EXISTS idx_maillog_create_time ON mail_logs (create_time);

