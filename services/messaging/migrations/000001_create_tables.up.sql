CREATE TABLE messaging_threads (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    subject TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'GENERAL',
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE messaging_messages (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    thread_id UUID NOT NULL REFERENCES messaging_threads(id),
    sender_id UUID NOT NULL,
    body TEXT NOT NULL,
    nonce TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE messaging_participants (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    thread_id UUID NOT NULL REFERENCES messaging_threads(id),
    participant_id UUID NOT NULL,
    role TEXT NOT NULL DEFAULT 'MEMBER',
    last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_messaging_threads_tenant ON messaging_threads(tenant_id, branch_id);
CREATE INDEX idx_messaging_threads_branch ON messaging_threads(branch_id);
CREATE INDEX idx_messaging_participants_participant ON messaging_participants(participant_id);
CREATE INDEX idx_messaging_messages_thread ON messaging_messages(thread_id);
CREATE INDEX idx_messaging_messages_tenant ON messaging_messages(tenant_id);
CREATE INDEX idx_messaging_participants_tenant ON messaging_participants(tenant_id);
