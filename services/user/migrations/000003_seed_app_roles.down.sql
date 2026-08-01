-- Remove role assignments and seeded users
DELETE FROM user_role_assignments
WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@clinic.com');

DELETE FROM user_profiles
WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@clinic.com');

DELETE FROM users WHERE email LIKE '%@clinic.com';

DELETE FROM user_roles WHERE name IN ('super_admin', 'app_admin', 'support', 'user');
