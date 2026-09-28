-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS group_permissions;
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;
