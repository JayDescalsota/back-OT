ALTER TABLE guardians ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE patient_guardians ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
CREATE INDEX idx_guardians_active ON guardians (is_active);
CREATE INDEX idx_patient_guardians_active ON patient_guardians (is_active);
