DROP INDEX IF EXISTS idx_patients_active;
ALTER TABLE patients DROP COLUMN IF EXISTS is_active;
