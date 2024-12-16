-- create permissions table
CREATE TABLE IF NOT EXISTS permissions (
  id bigserial PRIMARY KEY,
  code text NOT NULL
);

-- create join table for users and permission
-- make both fields primary key so that the same pair can't be duplicated
CREATE TABLE IF NOT EXISTS users_permissions (
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
  PRIMARY KEY(user_id, permission_id)
);

-- add two permissions to table
INSERT INTO permissions (code)
VALUES 
  ('movies:read'),
  ('movies:write');
