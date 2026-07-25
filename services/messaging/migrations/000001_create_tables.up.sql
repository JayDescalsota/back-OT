CREATE TABLE message_thread (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    type TEXT NOT NULL DEFAULT 'GENERAL',
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    created_action TEXT NOT NULL DEFAULT '',
    updated_by UUID,
    updated_action TEXT NOT NULL DEFAULT ''
);

CREATE TABLE message (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    thread_id UUID NOT NULL REFERENCES message_thread(id),
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

CREATE TABLE message_participant (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    thread_id UUID NOT NULL REFERENCES message_thread(id),
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

CREATE INDEX idx_message_thread_tenant ON message_thread(tenant_id, branch_id);
CREATE INDEX idx_message_thread_branch ON message_thread(branch_id);
CREATE INDEX idx_message_thread_participant ON message_participant(participant_id);
CREATE INDEX idx_message_thread_id ON message(thread_id);
CREATE INDEX idx_message_tenant ON message(tenant_id);
CREATE INDEX idx_message_participant_tenant ON message_participant(tenant_id);
