-- Remove user02@clinic.com
DELETE FROM sessions WHERE user_id = 'b1b2c3d4-1003-4000-8000-000000000003';
DELETE FROM user_branch_assignments WHERE user_id = 'b1b2c3d4-1003-4000-8000-000000000003';
DELETE FROM users WHERE id = 'b1b2c3d4-1003-4000-8000-000000000003';
