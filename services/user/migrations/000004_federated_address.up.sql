ALTER TABLE user_profiles ADD COLUMN address_id UUID;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS street_address;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS barangay;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS city;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS province;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS zip;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS country;
