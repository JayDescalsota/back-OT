CREATE TABLE tenant_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL,
    baranggay TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    zip_code TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT 'Philippines',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT,
    created_action TEXT NOT NULL DEFAULT 'CREATE',
    updated_by TEXT,
    updated_action TEXT NOT NULL DEFAULT 'CREATE'
);

ALTER TABLE tenant_branches ADD COLUMN address_id UUID REFERENCES tenant_addresses(id);
