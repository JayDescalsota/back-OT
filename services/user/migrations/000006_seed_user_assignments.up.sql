INSERT INTO users (id, email, password_hash, is_active, is_validated, created_at, updated_at)
VALUES
  ('b1b2c3d4-1001-4000-8000-000000000001', 'admin@clinic.com', '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
  ('b1b2c3d4-1002-4000-8000-000000000002', 'user@clinic.com', '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Assign admin@clinic.com as admin in Branch A1, therapist in Branch B1
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
FROM (SELECT id FROM users WHERE email = 'admin@clinic.com') u
CROSS JOIN (
  SELECT b.id, b.tenant_id FROM branches b WHERE b.id = 'b1b2c3d4-0001-4000-8000-000000000001'
  UNION ALL
  SELECT b.id, b.tenant_id FROM branches b WHERE b.id = 'b1b2c3d4-0003-4000-8000-000000000003'
) b
CROSS JOIN LATERAL (
  SELECT r.id FROM roles r WHERE r.name = CASE
    WHEN b.id = 'b1b2c3d4-0001-4000-8000-000000000001' THEN 'admin'
    WHEN b.id = 'b1b2c3d4-0003-4000-8000-000000000003' THEN 'therapist'
  END
) r
ON CONFLICT (user_id, branch_id, tenant_id) DO NOTHING;

-- Seed user@clinic.com as therapist in Branch A2, front_desk in Branch B2
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
FROM (SELECT id FROM users WHERE email = 'user@clinic.com') u
CROSS JOIN (
  SELECT b.id, b.tenant_id FROM branches b WHERE b.id = 'b1b2c3d4-0002-4000-8000-000000000002'
  UNION ALL
  SELECT b.id, b.tenant_id FROM branches b WHERE b.id = 'b1b2c3d4-0004-4000-8000-000000000004'
) b
CROSS JOIN LATERAL (
  SELECT r.id FROM roles r WHERE r.name = CASE
    WHEN b.id = 'b1b2c3d4-0002-4000-8000-000000000002' THEN 'therapist'
    WHEN b.id = 'b1b2c3d4-0004-4000-8000-000000000004' THEN 'front_desk'
  END
) r
ON CONFLICT (user_id, branch_id, tenant_id) DO NOTHING;
