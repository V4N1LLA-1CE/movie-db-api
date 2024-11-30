-- Add temporary integer column
ALTER TABLE movies ADD COLUMN version_int INTEGER DEFAULT 1;

-- Drop the UUID column and rename the integer column
ALTER TABLE movies DROP COLUMN version;
ALTER TABLE movies RENAME COLUMN version_int TO version;

-- Make version non-null
ALTER TABLE movies ALTER COLUMN version SET NOT NULL;
