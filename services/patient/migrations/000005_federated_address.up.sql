ALTER TABLE patients ADD COLUMN address_id UUID;
ALTER TABLE patient_guardian_profiles ADD COLUMN address_id UUID;

DROP TABLE IF EXISTS patient_addresses;
