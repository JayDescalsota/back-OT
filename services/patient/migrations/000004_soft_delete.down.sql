DROP INDEX IF EXISTS idx_guardian_links_active;
DROP INDEX IF EXISTS idx_guardian_profiles_active;
ALTER TABLE patient_guardian_links DROP COLUMN IF EXISTS is_active;
ALTER TABLE patient_guardian_profiles DROP COLUMN IF EXISTS is_active;
