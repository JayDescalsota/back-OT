ALTER TABLE user_profiles ADD COLUMN street_address TEXT;
ALTER TABLE user_profiles ADD COLUMN barangay TEXT;
ALTER TABLE user_profiles ADD COLUMN city TEXT;
ALTER TABLE user_profiles ADD COLUMN province TEXT;
ALTER TABLE user_profiles ADD COLUMN zip TEXT;
ALTER TABLE user_profiles ADD COLUMN country TEXT NOT NULL DEFAULT 'PH';
ALTER TABLE user_profiles DROP COLUMN IF EXISTS address_id;
