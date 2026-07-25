ALTER TABLE patients DROP COLUMN IF EXISTS address_id;
ALTER TABLE guardians DROP COLUMN IF EXISTS address_id;

CREATE TABLE IF NOT EXISTS patient_addresses (
    patient_id UUID PRIMARY KEY REFERENCES patients(id),
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
