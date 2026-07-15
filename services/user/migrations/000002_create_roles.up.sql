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

INSERT INTO roles (id, name, description) VALUES
    (gen_random_uuid(), 'admin', 'System administrator with full access'),
    (gen_random_uuid(), 'therapist', 'Occupational therapist'),
    (gen_random_uuid(), 'guardian', 'Patient guardian/parent'),
    (gen_random_uuid(), 'assistant_ot', 'Assistant occupational therapist'),
    (gen_random_uuid(), 'front_desk', 'Front desk staff'),
    (gen_random_uuid(), 'hr_manager', 'HR manager'),
    (gen_random_uuid(), 'executive', 'Executive leadership');
