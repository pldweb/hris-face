ALTER TABLE employees ADD COLUMN annual_leave_quota int NOT NULL DEFAULT 12;

CREATE TYPE leave_type AS ENUM ('annual', 'sick', 'permit');

CREATE TABLE leave_requests (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id  uuid NOT NULL REFERENCES employees(id),
    type         leave_type NOT NULL,
    start_date   date NOT NULL,
    end_date     date NOT NULL,
    days_count   int NOT NULL,
    reason       text NOT NULL,
    status       text NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    reviewed_by  uuid REFERENCES users(id),
    reviewed_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT leave_dates_valid CHECK (end_date >= start_date)
);
CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id, start_date DESC);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
