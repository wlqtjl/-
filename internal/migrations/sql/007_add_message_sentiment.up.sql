-- Add sentiment column to messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS sentiment TEXT;
