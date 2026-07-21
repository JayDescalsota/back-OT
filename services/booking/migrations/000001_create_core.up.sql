CREATE TABLE branch_hour (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    recurrence_rule TEXT NOT NULL,
    open_time TIMESTAMPTZ NOT NULL,
    close_time TIMESTAMPTZ NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE practitioner_branch (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    practitioner_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    role TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE practitioner_availability (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    practitioner_id UUID NOT NULL,
    recurrence_rule TEXT,
    available_from TIMESTAMPTZ NOT NULL,
    available_to TIMESTAMPTZ NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE schedule_template (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    practitioner_id UUID,
    recurrence_rule TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    slot_duration INT NOT NULL DEFAULT 30,
    buffer_duration INT NOT NULL DEFAULT 0,
    capacity INT NOT NULL DEFAULT 1,
    booked_count INT NOT NULL DEFAULT 0,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE appointment_slot (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    schedule_template_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    practitioner_id UUID,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    capacity INT NOT NULL DEFAULT 1,
    booked_count INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'AVAILABLE',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE appointment (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    appointment_slot_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    patient_id UUID,
    practitioner_id UUID,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    notes TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE schedule_exception (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    schedule_template_id UUID,
    appointment_slot_id UUID,
    practitioner_id UUID,
    type TEXT NOT NULL DEFAULT 'OVERRIDE',
    reason TEXT NOT NULL DEFAULT '',
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);
