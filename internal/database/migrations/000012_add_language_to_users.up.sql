-- Add language column to users table
ALTER TABLE users
ADD COLUMN language VARCHAR(10) NOT NULL DEFAULT 'en';

COMMENT ON COLUMN users.language IS 'User preferred language (en, ru, etc.)';
