DROP INDEX IF EXISTS idx_patient_guardians_active;
DROP INDEX IF EXISTS idx_guardians_active;
ALTER TABLE patient_guardians DROP COLUMN IF EXISTS is_active;
ALTER TABLE guardians DROP COLUMN IF EXISTS is_active;
