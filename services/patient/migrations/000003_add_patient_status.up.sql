ALTER TABLE patients ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
CREATE INDEX idx_patients_active ON patients (is_active);
