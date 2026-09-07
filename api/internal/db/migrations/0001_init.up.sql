-- Initial schema for HRIS face-recognition attendance.
-- Scale target: < 200 employees, 1-2 locations (see docs/PRD.md section 4).
-- No vector index: sequential scan over ~1,000 embeddings is sub-5ms at this scale.

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM ('employee', 'manager', 'hr', 'superadmin');

CREATE TABLE departments (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL UNIQUE,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE work_locations (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE work_schedules (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name              text NOT NULL,
    start_time        time NOT NULL,
    end_time          time NOT NULL,
    late_tolerance_minutes int NOT NULL DEFAULT 0,
    work_days         smallint[] NOT NULL DEFAULT '{1,2,3,4,5}', -- 1=Mon .. 7=Sun
    created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email          citext NOT NULL UNIQUE,
    password_hash  text NOT NULL,
    role           user_role NOT NULL DEFAULT 'employee',
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE employees (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nik              text NOT NULL UNIQUE,
    full_name        text NOT NULL,
    department_id    uuid REFERENCES departments(id),
    location_id      uuid REFERENCES work_locations(id),
    schedule_id      uuid REFERENCES work_schedules(id),
    manager_id       uuid REFERENCES employees(id),
    allow_remote     boolean NOT NULL DEFAULT false,
    hired_at         date,
    status           text NOT NULL DEFAULT 'pending_enrollment', -- pending_enrollment | active | inactive
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE face_embeddings (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    embedding     vector(512) NOT NULL,
    quality_score real NOT NULL,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_face_embeddings_employee ON face_embeddings(employee_id) WHERE is_active;

CREATE TABLE devices (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    device_key    text NOT NULL UNIQUE, -- cookie-issued device id
    user_agent    text,
    approved      boolean NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE attendances (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    uuid NOT NULL REFERENCES employees(id),
    type           text NOT NULL, -- check_in | check_out
    occurred_at    timestamptz NOT NULL DEFAULT now(), -- server time, never client
    similarity     real,
    liveness_score real,
    ip_address     inet,
    lat            double precision,
    lng            double precision,
    device_id      uuid REFERENCES devices(id),
    photo_path     text,
    status         text NOT NULL, -- on_time | late | early_leave | absent
    source         text NOT NULL DEFAULT 'face', -- face | manual_correction
    low_confidence boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_attendances_employee_time ON attendances(employee_id, occurred_at DESC);

CREATE TABLE attendance_corrections (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    uuid NOT NULL REFERENCES employees(id),
    requested_type text NOT NULL,
    requested_time timestamptz NOT NULL,
    reason         text NOT NULL,
    status         text NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    reviewed_by    uuid REFERENCES users(id),
    reviewed_at    timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    uuid REFERENCES users(id),
    action      text NOT NULL,
    entity      text NOT NULL,
    entity_id   uuid,
    before_data jsonb,
    after_data  jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);
