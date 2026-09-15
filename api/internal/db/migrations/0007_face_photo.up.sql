-- Path to the reference photo saved when an employee registers/re-registers
-- their face, so HR can see who was actually enrolled (docs request: "bisa
-- dilihat foto ketika daftar wajahnya"). Distinct from attendances.photo_path
-- (day-by-day evidence with 30-day retention) -- this one persists for as
-- long as the employee record does, and is overwritten on re-enroll.
ALTER TABLE employees ADD COLUMN face_photo_path text;
