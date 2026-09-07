-- Attendance status was hardcoded to 'on_time' because nothing ever read
-- work_schedules and employees had no schedule attached. Seed one sensible
-- default and attach it, so late/early_leave are computable out of the box.

INSERT INTO work_schedules (name, start_time, end_time, late_tolerance_minutes, work_days)
VALUES ('Reguler 08:00-17:00', '08:00', '17:00', 15, '{1,2,3,4,5}')
ON CONFLICT DO NOTHING;

UPDATE employees
SET schedule_id = (SELECT id FROM work_schedules ORDER BY created_at LIMIT 1)
WHERE schedule_id IS NULL;
