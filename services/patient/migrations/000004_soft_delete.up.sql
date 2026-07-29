ALTER TABLE patient_guardian_profiles ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE patient_guardian_links ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
CREATE INDEX idx_guardian_profiles_active ON patient_guardian_profiles (is_active);
CREATE INDEX idx_guardian_links_active ON patient_guardian_links (is_active);
