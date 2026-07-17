DELETE FROM user_app_roles
WHERE user_id IN (
    SELECT id FROM users WHERE email IN ('superadmin@clinic.com')
);

DELETE FROM users WHERE email IN (
    'superadmin@clinic.com', 'admin@clinic.com',
    'user01@clinic.com', 'user02@clinic.com'
);

DELETE FROM app_roles WHERE name IN ('super_admin', 'app_admin', 'support');
