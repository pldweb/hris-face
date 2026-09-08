-- An HR/superadmin account has no employees row, so it had nowhere to store a
-- name: the header fell back to the literal word "Admin" for everyone. Give
-- users their own display name for exactly those account-only cases; employees
-- keep employees.full_name as the canonical HR record.
ALTER TABLE users ADD COLUMN display_name text;
