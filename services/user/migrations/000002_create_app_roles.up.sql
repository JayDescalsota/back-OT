CREATE TABLE app_roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    created_action TEXT,
    updated_at TIMESTAMPTZ,
    updated_by UUID REFERENCES users(id),
    updated_action TEXT
);

CREATE TABLE user_app_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_role_id INT NOT NULL REFERENCES app_roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    updated_by UUID REFERENCES users(id),
    updated_action TEXT,
    PRIMARY KEY (user_id, app_role_id)
);

INSERT INTO app_roles (name, description) VALUES
    ('super_admin', 'Full cross-tenant system access'),
    ('app_admin', 'Manage tenants, users, and system configuration'),
    ('support', 'View system health, impersonate users for troubleshooting'),
    ('user', 'Regular tenant user');

INSERT INTO users (id, email, name, password_hash, is_active, is_validated, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'superadmin@clinic.com', 'Super Admin', '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000002', 'admin@clinic.com',     'Admin',        '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000003', 'user01@clinic.com',    'User 01',      '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000004', 'user02@clinic.com',    'User 02',      '$2a$10$qQalizyEz8LhRmGUZxyAp.A.1MUDAFmNERqhq/AIUxHfxQ2HsqAp2', true, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_app_roles (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN app_roles ar
WHERE u.email = 'superadmin@clinic.com' AND ar.name = 'super_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_app_roles (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN app_roles ar
WHERE u.email = 'admin@clinic.com' AND ar.name = 'app_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_app_roles (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN app_roles ar
WHERE u.email = 'user01@clinic.com' AND ar.name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO user_app_roles (user_id, app_role_id, assigned_at)
SELECT u.id, ar.id, NOW()
FROM users u CROSS JOIN app_roles ar
WHERE u.email = 'user02@clinic.com' AND ar.name = 'user'
ON CONFLICT DO NOTHING;
