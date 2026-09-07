DROP INDEX IF EXISTS idx_leave_requests_status;
DROP INDEX IF EXISTS idx_leave_requests_employee;
DROP TABLE IF EXISTS leave_requests;
DROP TYPE IF EXISTS leave_type;
ALTER TABLE employees DROP COLUMN annual_leave_quota;
