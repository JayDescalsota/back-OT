-- Seed tenants
INSERT INTO tenants (id, name, created_at, updated_at)
VALUES
  ('a1b2c3d4-0001-4000-8000-000000000001', 'Clinic A', NOW(), NOW()),
  ('a1b2c3d4-0002-4000-8000-000000000002', 'Clinic B', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed branches for Clinic A
INSERT INTO branches (id, tenant_id, name, created_at, updated_at)
VALUES
  ('b1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'Branch A1', NOW(), NOW()),
  ('b1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'Branch A2', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed branches for Clinic B
INSERT INTO branches (id, tenant_id, name, created_at, updated_at)
VALUES
  ('b1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0002-4000-8000-000000000002', 'Branch B1', NOW(), NOW()),
  ('b1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'Branch B2', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
