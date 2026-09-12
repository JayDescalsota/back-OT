-- Seed app roles
INSERT INTO user_roles (name, description) VALUES
    ('super_admin', 'Full cross-tenant system access'),
    ('app_admin', 'Manage tenants, users, and system configuration'),
    ('support', 'View system health, impersonate users for troubleshooting'),
    ('user', 'Regular tenant user')
ON CONFLICT (name) DO NOTHING;

-- Seed users
INSERT INTO users (id, email, password_hash, is_active, is_validated, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'superadmin@clinic.com', '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000002', 'admin@clinic.com',     '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000003', 'user01@clinic.com',    '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000004', 'user02@clinic.com',    '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000005', 'staff01@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000006', 'staff02@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000007', 'staff03@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000008', 'staff04@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000009', 'staff05@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000010', 'staff06@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000011', 'staff07@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000012', 'staff08@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000013', 'staff09@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000014', 'staff10@clinic.com',   '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Seed user_profiles for seeded users (split name into first/last by first space)
INSERT INTO user_profiles (user_id, first_name, last_name)
SELECT id,
       CASE WHEN POSITION(' ' IN COALESCE(name, email)) > 0 THEN LEFT(COALESCE(name, email), POSITION(' ' IN COALESCE(name, email)) - 1) ELSE COALESCE(name, email) END,
       CASE WHEN POSITION(' ' IN COALESCE(name, email)) > 0 THEN SUBSTRING(COALESCE(name, email) FROM POSITION(' ' IN COALESCE(name, email)) + 1) ELSE NULL END
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- Assign roles
INSERT INTO user_role_assignments (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN user_roles ar
WHERE u.email = 'superadmin@clinic.com' AND ar.name = 'super_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_role_assignments (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN user_roles ar
WHERE u.email = 'admin@clinic.com' AND ar.name = 'app_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_role_assignments (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN user_roles ar
WHERE u.email = 'user01@clinic.com' AND ar.name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO user_role_assignments (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN user_roles ar
WHERE u.email = 'user02@clinic.com' AND ar.name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO user_role_assignments (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN user_roles ar
WHERE u.email LIKE 'staff%@clinic.com' AND ar.name = 'user'
ON CONFLICT DO NOTHING;

-- Update users.name with real full names
UPDATE users SET name = 'Juan Carlos Dela Cruz' WHERE email = 'superadmin@clinic.com';
UPDATE users SET name = 'Maria Angelica Santos' WHERE email = 'admin@clinic.com';
UPDATE users SET name = 'Rafael Mendoza' WHERE email = 'user01@clinic.com';
UPDATE users SET name = 'Sophia Nicole Reyes' WHERE email = 'user02@clinic.com';
UPDATE users SET name = 'Miguel Angelo Torres' WHERE email = 'staff01@clinic.com';
UPDATE users SET name = 'Isabella Grace Flores' WHERE email = 'staff02@clinic.com';
UPDATE users SET name = 'Gabriel Dominic Cruz' WHERE email = 'staff03@clinic.com';
UPDATE users SET name = 'Camille Anne Bautista' WHERE email = 'staff04@clinic.com';
UPDATE users SET name = 'Andres Bonifacio Garcia' WHERE email = 'staff05@clinic.com';
UPDATE users SET name = 'Lara Mae Villanueva' WHERE email = 'staff06@clinic.com';
UPDATE users SET name = 'Nathaniel James Aquino' WHERE email = 'staff07@clinic.com';
UPDATE users SET name = 'Patricia Marie Navarro' WHERE email = 'staff08@clinic.com';
UPDATE users SET name = 'Christian Lawrence Ramos' WHERE email = 'staff09@clinic.com';
UPDATE users SET name = 'Angela Bea Domingo' WHERE email = 'staff10@clinic.com';

-- Backfill first/last names for existing profiles
UPDATE user_profiles
SET first_name = SPLIT_PART(u.name, ' ', 1),
    last_name = SPLIT_PART(u.name, ' ', 2)
FROM users u
WHERE user_profiles.user_id = u.id;

-- Seed complete user_profiles
INSERT INTO user_profiles (user_id, first_name, last_name, middle_name, suffix, title, phone, mobile, date_of_birth, gender, timezone, preferred_language, emergency_contact_name, emergency_contact_phone, emergency_contact_relation, notes)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Juan', 'Dela Cruz', 'Carlos', NULL, 'Dr.', '+63 2 8123 4567', '+63 917 123 4567', '1980-03-15', 'male', 'Asia/Manila', 'en', 'Maria Dela Cruz', '+63 917 765 4321', 'Spouse', 'Platform super administrator'),
    ('00000000-0000-0000-0000-000000000002', 'Maria', 'Santos', 'Angelica', NULL, 'Dr.', '+63 2 8765 4321', '+63 918 234 5678', '1985-07-22', 'female', 'Asia/Manila', 'en', 'Jose Santos', '+63 918 876 5432', 'Spouse', 'Clinic administrator'),
    ('00000000-0000-0000-0000-000000000003', 'Rafael', 'Mendoza', NULL, NULL, 'Dr.', '+63 2 8456 7890', '+63 919 345 6789', '1988-11-02', 'male', 'Asia/Manila', 'en', 'Luz Mendoza', '+63 919 987 6543', 'Mother', 'Occupational Therapist - Pediatrics'),
    ('00000000-0000-0000-0000-000000000004', 'Sophia', 'Reyes', 'Nicole', NULL, 'Dr.', '+63 2 8541 9632', '+63 920 456 7890', '1992-04-18', 'female', 'Asia/Manila', 'en', 'Ricardo Reyes', '+63 920 123 0987', 'Father', 'Occupational Therapist - Adults'),
    ('00000000-0000-0000-0000-000000000005', 'Miguel', 'Torres', 'Angelo', NULL, 'Dr.', '+63 2 8965 7410', '+63 921 567 8901', '1986-09-30', 'male', 'Asia/Manila', 'en', 'Elena Torres', '+63 921 111 2222', 'Mother', 'Occupational Therapist - Hand Therapy'),
    ('00000000-0000-0000-0000-000000000006', 'Isabella', 'Flores', 'Grace', NULL, NULL, '+63 2 8765 1245', '+63 922 678 9012', '1994-01-25', 'female', 'Asia/Manila', 'en', 'Antonio Flores', '+63 922 333 4444', 'Father', 'Assistant Occupational Therapist'),
    ('00000000-0000-0000-0000-000000000007', 'Gabriel', 'Cruz', 'Dominic', NULL, NULL, '+63 2 8452 3698', '+63 923 789 0123', '1990-06-14', 'male', 'Asia/Manila', 'en', 'Carmen Cruz', '+63 923 555 6666', 'Mother', 'Reception staff'),
    ('00000000-0000-0000-0000-000000000008', 'Camille', 'Bautista', 'Anne', NULL, NULL, '+63 2 8123 9874', '+63 924 890 1234', '1995-12-05', 'female', 'Asia/Manila', 'en', 'Lourdes Bautista', '+63 924 777 8888', 'Mother', 'Billing staff'),
    ('00000000-0000-0000-0000-000000000009', 'Andres', 'Garcia', 'Bonifacio', NULL, NULL, '+63 2 8654 7891', '+63 925 901 2345', '1983-08-09', 'male', 'Asia/Manila', 'en', 'Rosario Garcia', '+63 925 999 0000', 'Spouse', 'Logistics staff'),
    ('00000000-0000-0000-0000-000000000010', 'Lara', 'Villanueva', 'Mae', NULL, 'Dr.', '+63 2 8521 4763', '+63 926 012 3456', '1991-02-17', 'female', 'Asia/Manila', 'en', 'Ben Villanueva', '+63 926 111 2222', 'Spouse', 'Occupational Therapist - Pediatrics'),
    ('00000000-0000-0000-0000-000000000011', 'Nathaniel', 'Aquino', 'James', NULL, NULL, '+63 2 8345 6789', '+63 927 123 4567', '1988-05-23', 'male', 'Asia/Manila', 'en', 'Teresa Aquino', '+63 927 333 4444', 'Mother', 'Assistant Occupational Therapist'),
    ('00000000-0000-0000-0000-000000000012', 'Patricia', 'Navarro', 'Marie', NULL, NULL, '+63 2 8214 5863', '+63 928 234 5678', '1993-10-11', 'female', 'Asia/Manila', 'en', 'George Navarro', '+63 928 555 6666', 'Father', 'Front desk staff'),
    ('00000000-0000-0000-0000-000000000013', 'Christian', 'Ramos', 'Lawrence', NULL, NULL, '+63 2 8567 4321', '+63 929 345 6789', '1987-03-28', 'male', 'Asia/Manila', 'en', 'Andrea Ramos', '+63 929 777 8888', 'Spouse', 'Maintenance staff'),
    ('00000000-0000-0000-0000-000000000014', 'Angela', 'Domingo', 'Bea', NULL, NULL, '+63 2 8485 9632', '+63 930 456 7890', '1996-07-19', 'female', 'Asia/Manila', 'en', 'Ramon Domingo', '+63 930 999 0000', 'Father', 'Records staff')
ON CONFLICT (user_id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    middle_name = EXCLUDED.middle_name,
    suffix = EXCLUDED.suffix,
    title = EXCLUDED.title,
    phone = EXCLUDED.phone,
    mobile = EXCLUDED.mobile,
    date_of_birth = EXCLUDED.date_of_birth,
    gender = EXCLUDED.gender,
    timezone = EXCLUDED.timezone,
    preferred_language = EXCLUDED.preferred_language,
    emergency_contact_name = EXCLUDED.emergency_contact_name,
    emergency_contact_phone = EXCLUDED.emergency_contact_phone,
    emergency_contact_relation = EXCLUDED.emergency_contact_relation,
    notes = EXCLUDED.notes,
    updated_at = NOW();

-- Seed practitioner profiles for therapists and assistants
INSERT INTO user_practitioner_profiles (user_id, license_number, license_state, npi_number, specialty, sub_specialty, qualifications, credentials, education, years_of_experience, bio, is_accepting_patients)
VALUES
    ('00000000-0000-0000-0000-000000000003', 'OT-2010-0042', 'NCR', '1043290123', 'Occupational Therapy', 'Pediatrics', ARRAY['BS Occupational Therapy', 'MS Occupational Therapy'], ARRAY['Registered Occupational Therapist', 'Certified Hand Therapist'], 'University of Santo Tomas, 2008', 14, 'Specializes in pediatric occupational therapy with focus on sensory integration and fine motor development.', true),
    ('00000000-0000-0000-0000-000000000004', 'OT-2014-0091', 'NCR', '1785340291', 'Occupational Therapy', 'Adult Rehabilitation', ARRAY['BS Occupational Therapy'], ARRAY['Registered Occupational Therapist'], 'University of the Philippines Manila, 2012', 10, 'Experienced in adult neurorehabilitation and post-stroke recovery programs.', true),
    ('00000000-0000-0000-0000-000000000005', 'OT-2008-0017', 'NCR', '1392857402', 'Occupational Therapy', 'Hand Therapy', ARRAY['BS Occupational Therapy', 'Certified Hand Therapy'], ARRAY['Registered Occupational Therapist', 'Certified Hand Therapist'], 'De La Salle University, 2006', 16, 'Hand therapy specialist with 16 years of clinical experience in upper extremity rehabilitation.', true),
    ('00000000-0000-0000-0000-000000000006', NULL, 'NCR', NULL, 'Occupational Therapy', NULL, ARRAY['BS Occupational Therapy'], ARRAY['Registered Occupational Therapist'], 'Centro Escolar University, 2016', 8, 'Assistant occupational therapist supporting pediatric and adult treatment sessions.', true),
    ('00000000-0000-0000-0000-000000000010', 'OT-2013-0076', 'NCR', '1592038471', 'Occupational Therapy', 'Pediatrics', ARRAY['BS Occupational Therapy'], ARRAY['Registered Occupational Therapist'], 'University of Santo Tomas, 2011', 12, 'Pediatric occupational therapist focusing on early intervention and autism spectrum support.', true),
    ('00000000-0000-0000-0000-000000000011', NULL, 'NCR', NULL, 'Occupational Therapy', NULL, ARRAY['BS Occupational Therapy'], ARRAY['Registered Occupational Therapist'], 'University of Santo Tomas, 2010', 14, 'Assistant occupational therapist with geriatric rehabilitation experience.', true)
ON CONFLICT (user_id) DO NOTHING;

