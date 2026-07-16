INSERT INTO users (id, email, name, password_hash, is_active, is_validated, is_super_admin, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'superadmin@clinic.com',
  'Super Admin',
  '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2',
  true,
  true,
  true,
  NOW(),
  NOW()
)
ON CONFLICT (email) DO NOTHING;
