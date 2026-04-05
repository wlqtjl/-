-- Add profile fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT '';
