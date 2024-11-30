-- add uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- add temp UUID column
ALTER TABLE movies ADD COLUMN version_uuid UUID DEFAULT uuid_generate_v4();

-- generate uuid for each movie
UPDATE movies SET version_uuid = uuid_generate_v4();

-- drop old version column and rename new one
ALTER TABLE movies DROP COLUMN version;
ALTER TABLE movies RENAME COLUMN version_uuid TO version;

-- make version non null
ALTER TABLE movies ALTER COLUMN version SET NOT NULL;
