-- Remove seeded practitioner profiles
DELETE FROM user_practitioner_profiles WHERE user_id IN (
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000006',
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000011'
);

-- Reset names and profile details
UPDATE users SET name = 'User' WHERE email LIKE '%@clinic.com';
UPDATE user_profiles
SET first_name = 'User',
    last_name = NULL,
    middle_name = NULL,
    suffix = NULL,
    title = NULL,
    phone = NULL,
    mobile = NULL,
    date_of_birth = NULL,
    gender = NULL,
    emergency_contact_name = NULL,
    emergency_contact_phone = NULL,
    emergency_contact_relation = NULL,
    notes = NULL,
    updated_at = NOW();
