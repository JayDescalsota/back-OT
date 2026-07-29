CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO apps (id, name, display_name, description) VALUES
    ('a0000000-0001-4000-8000-000000000001', 'clinic-web', 'Clinic', 'Main clinic web application'),
    ('a0000000-0002-4000-8000-000000000002', 'analytics-web', 'Analytics', 'Analytics web application'),
    ('a0000000-0003-4000-8000-000000000003', 'guardian-mobile', 'Guardian Mobile', 'Mobile app for patient guardians'),
    ('a0000000-0004-4000-8000-000000000004', 'therapist-mobile', 'Therapist Mobile', 'Mobile app for therapists')
ON CONFLICT (id) DO NOTHING;
