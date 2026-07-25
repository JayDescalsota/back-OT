ALTER TABLE patients ADD COLUMN address_id UUID;
ALTER TABLE guardians ADD COLUMN address_id UUID;

DROP TABLE IF EXISTS patient_addresses;
