-- Create user02@clinic.com for Tenant B access
INSERT INTO users (id, email, name, password_hash, is_active, is_validated, created_at, updated_at)
VALUES (
  'b1b2c3d4-1003-4000-8000-000000000003',
  'user02@clinic.com',
  'User 02',
  '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2',
  true,
  true,
  NOW(),
  NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Assign user02@clinic.com as therapist in all Tenant B branches
INSERT INTO user_branch_assignments (id, user_id, branch_id, tenant_id, role_id, assigned_by, assigned_at, is_active)
SELECT
  gen_random_uuid(),
  u.id,
  b.id,
  b.tenant_id,
  r.id,
  u.id,
  NOW(),
  true
FROM (SELECT id FROM users WHERE email = 'user02@clinic.com') u
CROSS JOIN branches b
CROSS JOIN LATERAL (
  SELECT r.id FROM roles r WHERE r.name = 'therapist'
) r
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT (user_id, branch_id, tenant_id) DO NOTHING;
