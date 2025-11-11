-- Add profile fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT;

-- Create index on full name for search
CREATE INDEX IF NOT EXISTS idx_users_full_name ON users (first_name, last_name);

-- Comment on new columns
COMMENT ON COLUMN users.first_name IS 'User first name';
COMMENT ON COLUMN users.last_name IS 'User last name';
COMMENT ON COLUMN users.phone IS 'User phone number';
COMMENT ON COLUMN users.bio IS 'User biography/description';
COMMENT ON COLUMN users.avatar IS 'User avatar URL or base64 data';
