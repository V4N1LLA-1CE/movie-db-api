-- add case insensitive extension
CREATE EXTENSION IF NOT EXISTS citext;

-- create users table
CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  password_hash bytea NOT NULL,
  activated bool NOT NULL,
  version UUID NOT NULL DEFAULT uuid_generate_v4()
);
