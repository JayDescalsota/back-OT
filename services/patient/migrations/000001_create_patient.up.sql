CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth TIMESTAMPTZ,
    gender TEXT NOT NULL,
    notes TEXT,
    height TEXT,
    weight TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE TABLE patient_guardian_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE TABLE patient_guardian_links (
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    guardian_id UUID NOT NULL REFERENCES patient_guardian_profiles(id) ON DELETE CASCADE,
    relationship TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT,
    PRIMARY KEY (patient_id, guardian_id)
);

CREATE TABLE patient_tags (
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT,
    PRIMARY KEY (patient_id, tag_id)
);

CREATE TABLE patient_addresses (
    patient_id UUID PRIMARY KEY REFERENCES patients(id) ON DELETE CASCADE,
    address TEXT NOT NULL,
    baranggay TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    zip_code TEXT NOT NULL,
    country TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE INDEX idx_patients_tenant_branch ON patients (tenant_id, branch_id);
CREATE INDEX idx_patients_branch ON patients (branch_id);
