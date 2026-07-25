-- Seed tenants
INSERT INTO tenants (id, name, slug, domain)
VALUES
    ('a1b2c3d4-0001-4000-8000-000000000001', 'Clinic A', 'clinic-a', 'localhost'),
    ('a1b2c3d4-0002-4000-8000-000000000002', 'Clinic B', 'clinic-b', 'localhost')
ON CONFLICT (id) DO NOTHING;

-- Seed branches
INSERT INTO branches (id, tenant_id, name)
VALUES
    ('b1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'Branch A1'),
    ('b1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'Branch A2'),
    ('b1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0002-4000-8000-000000000002', 'Branch B1'),
    ('b1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'Branch B2')
ON CONFLICT (id) DO NOTHING;

-- System roles (shared across all tenants and branches)
INSERT INTO tenant_roles (tenant_id, branch_id, name, description, is_system_role)
SELECT 
    b.tenant_id, 
    b.id, 
    r.name, 
    r.description, 
    true
FROM branches b
CROSS JOIN (
    VALUES 
        ('branch_admin', 'Branch administrator with full access'),
        ('therapist', 'Occupational therapist'),
        ('guardian', 'Patient guardian/parent'),
        ('assistant_ot', 'Assistant occupational therapist'),
        ('front_desk', 'Front desk staff'),
        ('hr_manager', 'HR manager'),
        ('executive', 'Executive leadership')
) AS r(name, description)
ON CONFLICT DO NOTHING;

-- Seed permissions
INSERT INTO tenant_permissions (tenant_id, branch_id, resource, action, scope, description)
SELECT 
    b.tenant_id, 
    b.id, 
    p.resource, 
    p.action, 
    p.scope, 
    p.description
FROM branches b
CROSS JOIN (
    VALUES
        ('patient', 'read', 'branch', 'View patient records'),
        ('patient', 'write', 'branch', 'Create/edit patient records'),
        ('patient', 'delete', 'branch', 'Delete patient records'),
        ('appointment', 'read', 'own', 'View own appointments'),
        ('appointment', 'read', 'branch', 'View branch appointments'),
        ('appointment', 'write', 'branch', 'Create/edit appointments'),
        ('soap_note', 'read', 'own', 'View own SOAP notes'),
        ('soap_note', 'read', 'branch', 'View branch SOAP notes'),
        ('soap_note', 'write', 'branch', 'Create/edit SOAP notes'),
        ('soap_note', 'delete', 'branch', 'Delete SOAP notes'),
        ('assessment', 'read', 'own', 'View own assessments'),
        ('assessment', 'write', 'branch', 'Create/edit assessments'),
        ('treatment_plan', 'read', 'own', 'View own treatment plans'),
        ('treatment_plan', 'write', 'branch', 'Create/edit treatment plans'),
        ('goal', 'read', 'own', 'View own progress goals'),
        ('goal', 'write', 'branch', 'Create/edit progress goals'),
        ('employee', 'read', 'branch', 'View employee records'),
        ('employee', 'write', 'tenant', 'Manage employee records'),
        ('billing', 'read', 'branch', 'View billing records'),
        ('billing', 'write', 'tenant', 'Manage billing'),
        ('referral', 'read', 'branch', 'View referral records'),
        ('referral', 'write', 'branch', 'Create/manage referrals'),
        ('message', 'read', 'own', 'View own messages'),
        ('message', 'write', 'own', 'Send messages'),
        ('report', 'read', 'tenant', 'View reports'),
        ('config', 'write', 'tenant', 'Manage tenant configuration')
) AS p(resource, action, scope, description)
ON CONFLICT DO NOTHING;

-- Assign admin@clinic.com as branch_admin in all branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT 
    '00000000-0000-0000-0000-000000000002', 
    b.id, 
    b.tenant_id, 
    r.id, 
    '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'branch_admin'
ON CONFLICT DO NOTHING;

-- Assign user01@clinic.com as therapist in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT 
    '00000000-0000-0000-0000-000000000003', 
    b.id, 
    b.tenant_id, 
    r.id, 
    '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'therapist'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- Assign user02@clinic.com as therapist in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT 
    '00000000-0000-0000-0000-000000000004', 
    b.id, 
    b.tenant_id, 
    r.id, 
    '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'therapist'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;

-- Clinic A staff assignments
-- Maria Lopez as therapist in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000005', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'therapist'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- Josefina Cruz as assistant_ot in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000006', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'assistant_ot'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- Angela Reyes as front_desk in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000007', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'front_desk'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- David Tan as hr_manager in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000008', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'hr_manager'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- Catherine Lim as executive in Clinic A branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000009', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'executive'
WHERE b.tenant_id = 'a1b2c3d4-0001-4000-8000-000000000001'
ON CONFLICT DO NOTHING;

-- Clinic B staff assignments
-- Miguel Santos as therapist in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000010', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'therapist'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;

-- Patricia Gomez as assistant_ot in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000011', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'assistant_ot'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;

-- Luis Hernandez as front_desk in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000012', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'front_desk'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;

-- Sofia Martinez as hr_manager in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000013', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'hr_manager'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;

-- Carlos Mendoza as executive in Clinic B branches
INSERT INTO tenant_user_assignments (user_id, branch_id, tenant_id, role_id, assigned_by)
SELECT '00000000-0000-0000-0000-000000000014', b.id, b.tenant_id, r.id, '00000000-0000-0000-0000-000000000002'
FROM branches b
JOIN tenant_roles r ON r.branch_id = b.id AND r.name = 'executive'
WHERE b.tenant_id = 'a1b2c3d4-0002-4000-8000-000000000002'
ON CONFLICT DO NOTHING;
