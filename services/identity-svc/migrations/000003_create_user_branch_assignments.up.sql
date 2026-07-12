CREATE TABLE user_branch_assignments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    branch_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    assigned_by UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (user_id, branch_id, tenant_id)
);

CREATE INDEX idx_assignments_user ON user_branch_assignments (user_id);
CREATE INDEX idx_assignments_tenant_branch ON user_branch_assignments (tenant_id, branch_id);
