CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'own'
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

-- Seed default roles
INSERT INTO roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Full access to all branch data'),
    ('00000000-0000-0000-0000-000000000002', 'therapist', 'Clinical access for therapists'),
    ('00000000-0000-0000-0000-000000000003', 'guardian', 'Read-only access to own children'),
    ('00000000-0000-0000-0000-000000000004', 'assistant_ot', 'Limited clinical access under supervision'),
    ('00000000-0000-0000-0000-000000000005', 'front_desk', 'Scheduling and patient intake'),
    ('00000000-0000-0000-0000-000000000006', 'hr_manager', 'HR and credential management'),
    ('00000000-0000-0000-0000-000000000007', 'executive', 'Cross-branch analytics and reports');
